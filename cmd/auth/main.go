package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/hex"
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
	"github.com/ssoready/ssoready/internal/pagetoken"
	"github.com/ssoready/ssoready/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

	config := struct {
		SentryDSN              string `conf:"sentry-dsn,noredact"`
		SentryEnvironment      string `conf:"sentry-environment,noredact"`
		ServeAddr              string `conf:"serve-addr,noredact"`
		DB                     string `conf:"db"`
		GlobalDefaultAuthURL   string `conf:"global-default-auth-url,noredact"`
		PageEncodingSecret     string `conf:"page-encoding-secret"`
		SAMLStateSigningKey    string `conf:"saml-state-signing-key"`
		OAuthIDTokenPrivateKey string `conf:"oauth-id-token-private-key"`
	}{
		ServeAddr:            "localhost:8080",
		DB:                   "postgres://postgres:password@localhost/postgres",
		GlobalDefaultAuthURL: "http://localhost:8080",
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

	idTokenPrivateKey, err := parseRSAPrivateKey(config.OAuthIDTokenPrivateKey)
	if err != nil {
		panic(fmt.Errorf("parse oauth idtoken private key: %w", err))
	}

	service := authservice.Service{Store: store_, OAuthIDTokenPrivateKey: idTokenPrivateKey}

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

func parseRSAPrivateKey(s string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(s))
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
