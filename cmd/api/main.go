package main

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/resend/resend-go/v2"
	"github.com/rs/cors"
	"github.com/ssoready/ssoready/internal/apiservice"
	"github.com/ssoready/ssoready/internal/appauth/appauthinterceptor"
	"github.com/ssoready/ssoready/internal/gen/ssoready/v1/ssoreadyv1connect"
	"github.com/ssoready/ssoready/internal/google"
	"github.com/ssoready/ssoready/internal/pagetoken"
	"github.com/ssoready/ssoready/internal/store"
)

func main() {
	db, err := pgxpool.New(context.Background(), "postgres://postgres:password@localhost/postgres")
	if err != nil {
		panic(err)
	}

	// todo populate from env
	store_ := store.New(store.NewStoreParams{
		DB:                   db,
		PageEncoder:          pagetoken.Encoder{Secret: [32]byte{}},
		GlobalDefaultAuthURL: "http://localhost:8080",
		SAMLStateSigningKey:  [32]byte{},
	})

	connectPath, connectHandler := ssoreadyv1connect.NewSSOReadyServiceHandler(
		&apiservice.Service{
			Store: store_,
			GoogleClient: &google.Client{
				HTTPClient:          http.DefaultClient,
				GoogleOAuthClientID: "171906208332-m8dg2p6av2f0aa7lliaj6oo0grct57p1.apps.googleusercontent.com",
			},
			ResendClient:              resend.NewClient("re_jEFaTdrh_EmCQcqC9PZU4SeKY21XWSpcB"),
			EmailChallengeFrom:        "onboarding@resend.dev",
			EmailVerificationEndpoint: "https://app.ssoready.com/verify-email",
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
	mux.Handle("/internal/connect/", http.StripPrefix("/internal/connect", connectMux))
	mux.Handle("/", transcoder)
	if err := http.ListenAndServe("localhost:8081", mux); err != nil {
		panic(err)
	}
}
