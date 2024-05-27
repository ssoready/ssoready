package main

import (
	"context"
	_ "embed"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"

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
		ServeAddr            string `conf:"serve-addr,noredact"`
		DB                   string `conf:"db"`
		GlobalDefaultAuthURL string `conf:"global-default-auth-url,noredact"`
		PageEncodingSecret   string `conf:"page-encoding-secret"`
		SAMLStateSigningKey  string `conf:"saml-state-signing-key"`
	}{
		ServeAddr:            "localhost:8080",
		DB:                   "postgres://postgres:password@localhost/postgres",
		GlobalDefaultAuthURL: "http://localhost:8080",
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

	service := authservice.Service{Store: store_}

	r := mux.NewRouter()

	r.HandleFunc("/internal/health", func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "health")
		w.WriteHeader(http.StatusOK)
	}).Methods("GET")

	r.PathPrefix("/").Handler(service.NewHandler())

	slog.Info("serve")
	if err := http.ListenAndServe(config.ServeAddr, r); err != nil {
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
