package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/google"
	"github.com/ssoready/ssoready/internal/store"
)

func (s *Service) SignIn(ctx context.Context, req *connect.Request[ssoreadyv1.SignInRequest]) (*connect.Response[ssoreadyv1.SignInResponse], error) {
	credRes, err := s.GoogleClient.ParseCredential(ctx, &google.ParseCredentialRequest{
		Credential: req.Msg.GoogleCredential,
	})

	if err != nil {
		return nil, fmt.Errorf("google: parse credential: %w", err)
	}

	createSessionRes, err := s.Store.CreateGoogleSession(ctx, &store.CreateGoogleSessionRequest{
		Email:        credRes.Email,
		HostedDomain: credRes.HostedDomain,
		DisplayName:  credRes.Name,
	})

	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(&ssoreadyv1.SignInResponse{
		SessionToken: createSessionRes.SessionToken,
	}), nil
}
