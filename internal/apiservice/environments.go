package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) ListEnvironments(ctx context.Context, req *connect.Request[ssoreadyv1.ListEnvironmentsRequest]) (*connect.Response[ssoreadyv1.ListEnvironmentsResponse], error) {
	res, err := s.Store.ListEnvironments(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) GetEnvironment(ctx context.Context, req *connect.Request[ssoreadyv1.GetEnvironmentRequest]) (*connect.Response[ssoreadyv1.Environment], error) {
	res, err := s.Store.GetEnvironment(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) UpdateEnvironment(ctx context.Context, req *connect.Request[ssoreadyv1.UpdateEnvironmentRequest]) (*connect.Response[ssoreadyv1.Environment], error) {
	res, err := s.Store.UpdateEnvironment(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}
