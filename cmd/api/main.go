package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/resend/resend-go/v2"
	"github.com/rs/cors"
	"github.com/segmentio/analytics-go/v3"
	"github.com/ssoready/conf"
	"github.com/ssoready/ssoready/internal/apiservice"
	"github.com/ssoready/ssoready/internal/appanalytics"
	"github.com/ssoready/ssoready/internal/authn/authninterceptor"
	"github.com/ssoready/ssoready/internal/gen/ssoready/v1/ssoreadyv1connect"
	"github.com/ssoready/ssoready/internal/google"
	"github.com/ssoready/ssoready/internal/microsoft"
	"github.com/ssoready/ssoready/internal/pagetoken"
	"github.com/ssoready/ssoready/internal/secretload"
	"github.com/ssoready/ssoready/internal/sentryinterceptor"
	"github.com/ssoready/ssoready/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

	if err := secretload.Load(context.Background()); err != nil {
		panic(fmt.Errorf("load secrets: %w", err))
	}

	config := struct {
		SentryDSN                  string `conf:"sentry-dsn,noredact"`
		SentryEnvironment          string `conf:"sentry-environment,noredact"`
		ServeAddr                  string `conf:"serve-addr,noredact"`
		DB                         string `conf:"db"`
		GlobalDefaultAuthURL       string `conf:"global-default-auth-url,noredact"`
		PageEncodingSecret         string `conf:"page-encoding-secret"`
		SAMLStateSigningKey        string `conf:"saml-state-signing-key"`
		GoogleOAuthClientID        string `conf:"google-oauth-client-id,noredact"`
		MicrosoftOAuthClientID     string `conf:"microsoft-oauth-client-id,noredact"`
		MicrosoftOAuthClientSecret string `conf:"microsoft-oauth-client-secret"`
		MicrosoftOAuthRedirectURI  string `conf:"microsoft-oauth-redirect-uri,noredact"`
		ResendAPIKey               string `conf:"resend-api-key"`
		EmailChallengeFrom         string `conf:"email-challenge-from,noredact"`
		EmailVerificationEndpoint  string `conf:"email-verification-endpoint,noredact"`
		SegmentWriteKey            string `conf:"segment-write-key"`
	}{
		ServeAddr:                 "localhost:8081",
		DB:                        "postgres://postgres:password@localhost/postgres",
		GlobalDefaultAuthURL:      "http://localhost:8080",
		EmailChallengeFrom:        "onboarding@resend.dev",
		EmailVerificationEndpoint: "http://localhost:8082/verify-email",
	}

	conf.Load(&config)
	slog.Info("config", "config", conf.Redact(config))

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.SentryDSN,
		Environment:      config.SentryEnvironment,
		TracesSampleRate: 1.0,
	}); err != nil {
		panic(err)
	}

	db, err := pgxpool.New(context.Background(), config.DB)
	if err != nil {
		panic(err)
	}

	pageEncodingSecret, err := parseHexKey(config.PageEncodingSecret)
	if err != nil {
		panic(fmt.Errorf("parse page encoding secret: %w", err))
	}

	samlStateSigningKey, err := parseHexKey(config.SAMLStateSigningKey)
	if err != nil {
		panic(fmt.Errorf("parse saml state signing key: %w", err))
	}

	store_ := store.New(store.NewStoreParams{
		DB:                   db,
		PageEncoder:          pagetoken.Encoder{Secret: pageEncodingSecret},
		GlobalDefaultAuthURL: config.GlobalDefaultAuthURL,
		SAMLStateSigningKey:  samlStateSigningKey,
	})

	var analyticsClient analytics.Client = appanalytics.NoopClient{}
	if config.SegmentWriteKey != "" {
		analyticsClient = analytics.New(config.SegmentWriteKey)
	}

	connectPath, connectHandler := ssoreadyv1connect.NewSSOReadyServiceHandler(
		&apiservice.Service{
			Store: store_,
			GoogleClient: &google.Client{
				HTTPClient:          http.DefaultClient,
				GoogleOAuthClientID: config.GoogleOAuthClientID,
			},
			MicrosoftClient: &microsoft.Client{
				HTTPClient:                 http.DefaultClient,
				MicrosoftOAuthClientID:     config.MicrosoftOAuthClientID,
				MicrosoftOAuthClientSecret: config.MicrosoftOAuthClientSecret,
				MicrosoftOAuthRedirectURI:  config.MicrosoftOAuthRedirectURI,
			},
			ResendClient:              resend.NewClient(config.ResendAPIKey),
			EmailChallengeFrom:        config.EmailChallengeFrom,
			EmailVerificationEndpoint: config.EmailVerificationEndpoint,
			SAMLMetadataHTTPClient:    http.DefaultClient,
		},
		connect.WithInterceptors(
			sentryinterceptor.NewPreAuthentication(),
			authninterceptor.New(store_),
			sentryinterceptor.NewPostAuthentication(),
			appanalytics.NewInterceptor(analyticsClient),
		),
	)

	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})
	sentryConnectHandler := sentryHandler.Handle(connectHandler)

	service := vanguard.NewService(connectPath, sentryConnectHandler)
	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{service})
	if err != nil {
		panic(err)
	}

	connectMux := http.NewServeMux()
	connectMux.Handle(connectPath, cors.AllowAll().Handler(sentryConnectHandler))

	mux := http.NewServeMux()
	mux.Handle("/internal/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "health")
		w.WriteHeader(http.StatusOK)
	}))
	mux.Handle("/internal/connect/", http.StripPrefix("/internal/connect", connectMux))
	mux.Handle("/", transcoder)

	slog.Info("serve")
	if err := http.ListenAndServe(config.ServeAddr, mux); err != nil {
		panic(err)
	}
}

func parseHexKey(s string) ([32]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return [32]byte{}, err
	}

	if len(b) > 32 {
		return [32]byte{}, fmt.Errorf("key must encode 32 bytes")
	}

	var k [32]byte
	copy(k[:], b)
	return k, nil
}
