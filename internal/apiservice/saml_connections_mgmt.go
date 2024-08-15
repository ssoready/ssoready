package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) ListSAMLConnections(ctx context.Context, req *connect.Request[ssoreadyv1.ListSAMLConnectionsRequest]) (*connect.Response[ssoreadyv1.ListSAMLConnectionsResponse], error) {
	res, err := s.Store.ListSAMLConnections(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) GetSAMLConnection(ctx context.Context, req *connect.Request[ssoreadyv1.GetSAMLConnectionRequest]) (*connect.Response[ssoreadyv1.GetSAMLConnectionResponse], error) {
	res, err := s.Store.GetSAMLConnection(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) CreateSAMLConnection(ctx context.Context, req *connect.Request[ssoreadyv1.CreateSAMLConnectionRequest]) (*connect.Response[ssoreadyv1.CreateSAMLConnectionResponse], error) {
	res, err := s.Store.CreateSAMLConnection(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) UpdateSAMLConnection(ctx context.Context, req *connect.Request[ssoreadyv1.UpdateSAMLConnectionRequest]) (*connect.Response[ssoreadyv1.UpdateSAMLConnectionResponse], error) {
	res, err := s.Store.UpdateSAMLConnection(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}
