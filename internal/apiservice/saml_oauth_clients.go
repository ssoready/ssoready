package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/segmentio/analytics-go/v3"
	"github.com/ssoready/ssoready/internal/appanalytics"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) ListSAMLOAuthClients(ctx context.Context, req *connect.Request[ssoreadyv1.ListSAMLOAuthClientsRequest]) (*connect.Response[ssoreadyv1.ListSAMLOAuthClientsResponse], error) {
	res, err := s.Store.ListSAMLOAuthClients(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) GetSAMLOAuthClient(ctx context.Context, req *connect.Request[ssoreadyv1.GetSAMLOAuthClientRequest]) (*connect.Response[ssoreadyv1.SAMLOAuthClient], error) {
	res, err := s.Store.GetSAMLOAuthClient(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) CreateSAMLOAuthClient(ctx context.Context, req *connect.Request[ssoreadyv1.CreateSAMLOAuthClientRequest]) (*connect.Response[ssoreadyv1.SAMLOAuthClient], error) {
	res, err := s.Store.CreateSAMLOAuthClient(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if err := appanalytics.Track(ctx, "SAML OAuth Client Created", analytics.Properties{
		"environment_id": res.EnvironmentId,
		"api_key_id":     res.Id,
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) DeleteSAMLOAuthClient(ctx context.Context, req *connect.Request[ssoreadyv1.DeleteSAMLOAuthClientRequest]) (*connect.Response[emptypb.Empty], error) {
	res, err := s.Store.DeleteSAMLOAuthClient(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if err := appanalytics.Track(ctx, "SAML OAuth Client Deleted", analytics.Properties{
		"api_key_id": req.Msg.Id,
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
