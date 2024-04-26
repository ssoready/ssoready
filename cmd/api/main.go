package main

import (
	"net/http"

	"connectrpc.com/vanguard"
	"github.com/ssoready/ssoready/internal/gen/ssoready/v1/ssoreadyv1connect"
)

func main() {
	service := vanguard.NewService(
		ssoreadyv1connect.NewSSOReadyServiceHandler(
			ssoreadyv1connect.UnimplementedSSOReadyServiceHandler{},
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
