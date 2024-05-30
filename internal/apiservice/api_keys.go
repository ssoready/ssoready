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

func (s *Service) ListAPIKeys(ctx context.Context, req *connect.Request[ssoreadyv1.ListAPIKeysRequest]) (*connect.Response[ssoreadyv1.ListAPIKeysResponse], error) {
	res, err := s.Store.ListAPIKeys(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) GetAPIKey(ctx context.Context, req *connect.Request[ssoreadyv1.GetAPIKeyRequest]) (*connect.Response[ssoreadyv1.APIKey], error) {
	res, err := s.Store.GetAPIKey(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) CreateAPIKey(ctx context.Context, req *connect.Request[ssoreadyv1.CreateAPIKeyRequest]) (*connect.Response[ssoreadyv1.APIKey], error) {
	res, err := s.Store.CreateAPIKey(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if err := appanalytics.Track(ctx, "API Key Created", analytics.Properties{
		"environment_id": res.EnvironmentId,
		"api_key_id":     res.Id,
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) DeleteAPIKey(ctx context.Context, req *connect.Request[ssoreadyv1.DeleteAPIKeyRequest]) (*connect.Response[emptypb.Empty], error) {
	res, err := s.Store.DeleteAPIKey(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if err := appanalytics.Track(ctx, "API Key Deleted", analytics.Properties{
		"api_key_id": req.Msg.Id,
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
