package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssoready/conf"
	"github.com/ssoready/ssoready/internal/authservice"
	"github.com/ssoready/ssoready/internal/hexkey"
	"github.com/ssoready/ssoready/internal/pagetoken"
	"github.com/ssoready/ssoready/internal/secretload"
	"github.com/ssoready/ssoready/internal/store"
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
		BaseURL                      string `conf:"base-url,noredact"`
		PageEncodingValue            string `conf:"page-encoding-value"`
		SAMLStateSigningKey          string `conf:"saml-state-signing-key"`
		OAuthIDTokenPrivateKeyBase64 string `conf:"oauth-id-token-private-key-base64"`
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
		DB:                  db,
		PageEncoder:         pagetoken.Encoder{Secret: pageEncodingValue},
		DefaultAuthURL:      config.DefaultAuthURL,
		SAMLStateSigningKey: samlStateSigningKey,
	})

	idTokenPrivateKey, err := parseRSAPrivateKey(config.OAuthIDTokenPrivateKeyBase64)
	if err != nil {
		panic(fmt.Errorf("parse oauth idtoken private key: %w", err))
	}

	service := authservice.Service{
		Store:                  store_,
		BaseURL:                config.BaseURL,
		OAuthIDTokenPrivateKey: idTokenPrivateKey,
	}

	r := mux.NewRouter()

	r.HandleFunc("/internal/health", func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "health")
		w.WriteHeader(http.StatusOK)
	}).Methods("GET")

	r.PathPrefix("/").Handler(service.NewHandler())

	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})
	sentryMux := sentryHandler.Handle(r)

	logMux := logHTTP(sentryMux)

	slog.Info("serve")
	if err := http.ListenAndServe(config.ServeAddr, logMux); err != nil {
		panic(err)
	}
}

func logHTTP(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		slog.InfoContext(ctx, "http_request", "path", r.URL.Path)
		h.ServeHTTP(w, r)
	}
}

// parseRSAPrivateKey parses a JSON-encoded string containing a PKCS8 RSA private key.
func parseRSAPrivateKey(s string) (*rsa.PrivateKey, error) {
	if s == "" {
		slog.Warn("no oauth-id-token-private-key-base64 provided, generating new private key")

		k, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, fmt.Errorf("generate rsa key: %w", err)
		}
		return k, nil
	}

	// we base64-encode the string to avoid having ASCII newlines in the secret value, because such values do not play
	// nicely with e.g. AWS Secrets Manager.

	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid pem file")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not rsa private key")
	}

	return rsaKey, nil
}
