package main

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssoready/ssoready/internal/apiservice"
	"github.com/ssoready/ssoready/internal/appauth/appauthinterceptor"
	"github.com/ssoready/ssoready/internal/gen/ssoready/v1/ssoreadyv1connect"
	"github.com/ssoready/ssoready/internal/store"
)

func main() {
	db, err := pgxpool.New(context.Background(), "postgres://postgres:password@localhost/postgres")
	if err != nil {
		panic(err)
	}

	store_ := store.New(db)

	service := vanguard.NewService(
		ssoreadyv1connect.NewSSOReadyServiceHandler(
			&apiservice.Service{
				Store: store_,
			},
			connect.WithInterceptors(appauthinterceptor.New(store_)),
		),
	)
	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{service})
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", transcoder)
	if err := http.ListenAndServe("localhost:8081", mux); err != nil {
		panic(err)
	}
}
