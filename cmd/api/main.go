package main

import (
	"context"
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
	"github.com/ssoready/ssoready/internal/apistripewebhook"
	"github.com/ssoready/ssoready/internal/appanalytics"
	"github.com/ssoready/ssoready/internal/authn/authninterceptor"
	"github.com/ssoready/ssoready/internal/gen/ssoready/v1/ssoreadyv1connect"
	"github.com/ssoready/ssoready/internal/google"
	"github.com/ssoready/ssoready/internal/hexkey"
	"github.com/ssoready/ssoready/internal/microsoft"
	"github.com/ssoready/ssoready/internal/pagetoken"
	"github.com/ssoready/ssoready/internal/secretload"
	"github.com/ssoready/ssoready/internal/sentryinterceptor"
	"github.com/ssoready/ssoready/internal/store"
	stripe "github.com/stripe/stripe-go/v79/client"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

	if err := secretload.Load(context.Background()); err != nil {
		panic(fmt.Errorf("load secrets: %w", err))
	}

	config := struct {
		SentryDSN                    string `conf:"sentry-dsn,noredact"`
		SentryEnvironment            string `conf:"sentry-environment,noredact"`
		ServeAddr                    string `conf:"serve-addr,noredact"`
		DB                           string `conf:"db"`
		DefaultAuthURL               string `conf:"default-auth-url,noredact"`
		DefaultAdminSetupURL         string `conf:"default-admin-setup-url,noredact"`
		PageEncodingValue            string `conf:"page-encoding-value"`
		SAMLStateSigningKey          string `conf:"saml-state-signing-key"`
		GoogleOAuthClientID          string `conf:"google-oauth-client-id,noredact"`
		MicrosoftOAuthClientID       string `conf:"microsoft-oauth-client-id,noredact"`
		MicrosoftOAuthClientSecret   string `conf:"microsoft-oauth-client-secret"`
		MicrosoftOAuthRedirectURI    string `conf:"microsoft-oauth-redirect-uri,noredact"`
		ResendAPIKey                 string `conf:"resend-api-key"`
		EmailChallengeFrom           string `conf:"email-challenge-from,noredact"`
		EmailVerificationEndpoint    string `conf:"email-verification-endpoint,noredact"`
		SegmentWriteKey              string `conf:"segment-write-key"`
		StripeAPIKey                 string `conf:"stripe-api-key"`
		StripeCheckoutSuccessURL     string `conf:"stripe-checkout-success-url"`
		StripePriceIDProTier         string `conf:"stripe-price-id-pro-tier,noredact"`
		StripeWebhookEndpointSecret  string `conf:"stripe-webhook-endpoint-secret"`
		StripeBillingPortalReturnURL string `conf:"stripe-billing-portal-return-url,noredact"`
	}{
		PageEncodingValue: "0000000000000000000000000000000000000000000000000000000000000000",
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

	pageEncodingValue, err := hexkey.New(config.PageEncodingValue)
	if err != nil {
		panic(fmt.Errorf("parse page encoding secret: %w", err))
	}

	samlStateSigningKey, err := hexkey.New(config.SAMLStateSigningKey)
	if err != nil {
		panic(fmt.Errorf("parse saml state signing key: %w", err))
	}

	store_ := store.New(store.NewStoreParams{
		DB:                   db,
		PageEncoder:          pagetoken.Encoder{Secret: pageEncodingValue},
		DefaultAuthURL:       config.DefaultAuthURL,
		DefaultAdminSetupURL: config.DefaultAdminSetupURL,
		SAMLStateSigningKey:  samlStateSigningKey,
	})

	var analyticsClient analytics.Client = appanalytics.NoopClient{}
	if config.SegmentWriteKey != "" {
		analyticsClient = analytics.New(config.SegmentWriteKey)
	}

	stripeClient := &stripe.API{}
	stripeClient.Init(config.StripeAPIKey, nil)

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
			ResendClient:                        resend.NewClient(config.ResendAPIKey),
			EmailChallengeFrom:                  config.EmailChallengeFrom,
			EmailVerificationEndpoint:           config.EmailVerificationEndpoint,
			SAMLMetadataHTTPClient:              http.DefaultClient,
			StripeClient:                        stripeClient,
			StripeCheckoutSuccessURL:            config.StripeCheckoutSuccessURL,
			StripePriceIDProTier:                config.StripePriceIDProTier,
			StripeBillingPortalReturnURL:        config.StripeBillingPortalReturnURL,
			UnimplementedSSOReadyServiceHandler: ssoreadyv1connect.UnimplementedSSOReadyServiceHandler{},
		},
		connect.WithInterceptors(
			sentryinterceptor.NewPreAuthentication(),
			authninterceptor.New(store_),
			sentryinterceptor.NewPostAuthentication(),
			appanalytics.NewInterceptor(analyticsClient),
		),
	)

	stripeWebhookService := apistripewebhook.Service{
		Store:                store_,
		StripeClient:         stripeClient,
		StripeEndpointSecret: config.StripeWebhookEndpointSecret,
	}

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

	mux.Handle("/internal/stripe-webhook", stripeWebhookService.Handler())

	mux.Handle("/internal/connect/", http.StripPrefix("/internal/connect", connectMux))

	// hack workaround for camel-cased query params
	mux.Handle("/v1/scim/", publicAPICamelToSnake(transcoder))
	mux.Handle("/v1/organizations", publicAPICamelToSnake(transcoder))
	mux.Handle("/v1/saml-connections", publicAPICamelToSnake(transcoder))
	mux.Handle("/v1/scim-directories", publicAPICamelToSnake(transcoder))
	mux.Handle("/", transcoder)

	slog.Info("serve")
	if err := http.ListenAndServe(config.ServeAddr, mux); err != nil {
		panic(err)
	}
}

var publicAPICamelToSnakeMapping = map[string]string{
	"scimDirectoryId":        "scim_directory_id",
	"organizationId":         "organization_id",
	"organizationExternalId": "organization_external_id",
	"scimGroupId":            "scim_group_id",
	"pageToken":              "page_token",
}

// publicAPICamelToSnake converts public SCIM-related endpoint parameters from camel to snake.
//
// Workaround for: https://github.com/connectrpc/vanguard-go/issues/131
func publicAPICamelToSnake(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			q := r.URL.Query()
			for camel, snake := range publicAPICamelToSnakeMapping {
				if q.Has(camel) {
					v := q.Get(camel)
					q.Del(camel)
					q.Set(snake, v)
				}
			}

			r.URL.RawQuery = q.Encode()
		}

		h.ServeHTTP(w, r)
	})
}
