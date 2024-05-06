package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) ListOrganizations(ctx context.Context, req *connect.Request[ssoreadyv1.ListOrganizationsRequest]) (*connect.Response[ssoreadyv1.ListOrganizationsResponse], error) {
	res, err := s.Store.ListOrganizations(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) GetOrganization(ctx context.Context, req *connect.Request[ssoreadyv1.GetOrganizationRequest]) (*connect.Response[ssoreadyv1.Organization], error) {
	res, err := s.Store.GetOrganization(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) CreateOrganization(ctx context.Context, req *connect.Request[ssoreadyv1.CreateOrganizationRequest]) (*connect.Response[ssoreadyv1.Organization], error) {
	res, err := s.Store.CreateOrganization(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) UpdateOrganization(ctx context.Context, req *connect.Request[ssoreadyv1.UpdateOrganizationRequest]) (*connect.Response[ssoreadyv1.Organization], error) {
	res, err := s.Store.UpdateOrganization(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}
