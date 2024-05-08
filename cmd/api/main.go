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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/resend/resend-go/v2"
	"github.com/rs/cors"
	"github.com/ssoready/conf"
	"github.com/ssoready/ssoready/internal/apiservice"
	"github.com/ssoready/ssoready/internal/appauth/appauthinterceptor"
	"github.com/ssoready/ssoready/internal/gen/ssoready/v1/ssoreadyv1connect"
	"github.com/ssoready/ssoready/internal/google"
	"github.com/ssoready/ssoready/internal/pagetoken"
	"github.com/ssoready/ssoready/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

	config := struct {
		ServeAddr                 string `conf:"serve-addr,noredact"`
		DB                        string `conf:"db"`
		GlobalDefaultAuthURL      string `conf:"global-default-auth-url,noredact"`
		PageEncodingSecret        string `conf:"page-encoding-secret"`
		SAMLStateSigningKey       string `conf:"saml-state-signing-key"`
		GoogleOAuthClientID       string `conf:"google-oauth-client-id,noredact"`
		ResendAPIKey              string `conf:"resend-api-key"`
		EmailChallengeFrom        string `conf:"email-challenge-from,noredact"`
		EmailVerificationEndpoint string `conf:"email-verification-endpoint,noredact"`
	}{
		ServeAddr:                 "localhost:8081",
		DB:                        "postgres://postgres:password@localhost/postgres",
		GlobalDefaultAuthURL:      "http://localhost:8080",
		EmailChallengeFrom:        "onboarding@resend.dev",
		EmailVerificationEndpoint: "https://localhost:8082/verify-email",
	}

	conf.Load(&config)
	slog.Info("config", "config", conf.Redact(config))

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

	connectPath, connectHandler := ssoreadyv1connect.NewSSOReadyServiceHandler(
		&apiservice.Service{
			Store: store_,
			GoogleClient: &google.Client{
				HTTPClient:          http.DefaultClient,
				GoogleOAuthClientID: config.GoogleOAuthClientID,
			},
			ResendClient:              resend.NewClient(config.ResendAPIKey),
			EmailChallengeFrom:        config.EmailChallengeFrom,
			EmailVerificationEndpoint: config.EmailVerificationEndpoint,
		},
		connect.WithInterceptors(appauthinterceptor.New(store_)),
	)

	service := vanguard.NewService(connectPath, connectHandler)
	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{service})
	if err != nil {
		panic(err)
	}

	connectMux := http.NewServeMux()
	connectMux.Handle(connectPath, cors.AllowAll().Handler(connectHandler))

	mux := http.NewServeMux()
	mux.Handle("/internal/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	mux.Handle("/internal/connect/", http.StripPrefix("/internal/connect", connectMux))
	mux.Handle("/", transcoder)
	if err := http.ListenAndServe("localhost:8081", mux); err != nil {
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
