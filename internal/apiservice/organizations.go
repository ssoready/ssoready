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

func (s *Service) AppListOrganizations(ctx context.Context, req *connect.Request[ssoreadyv1.AppListOrganizationsRequest]) (*connect.Response[ssoreadyv1.AppListOrganizationsResponse], error) {
	res, err := s.Store.AppListOrganizations(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppGetOrganization(ctx context.Context, req *connect.Request[ssoreadyv1.AppGetOrganizationRequest]) (*connect.Response[ssoreadyv1.Organization], error) {
	res, err := s.Store.AppGetOrganization(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppCreateOrganization(ctx context.Context, req *connect.Request[ssoreadyv1.AppCreateOrganizationRequest]) (*connect.Response[ssoreadyv1.Organization], error) {
	res, err := s.Store.AppCreateOrganization(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if err := appanalytics.Track(ctx, "Organization Created", analytics.Properties{
		"environment_id":  res.EnvironmentId,
		"organization_id": res.Id,
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppUpdateOrganization(ctx context.Context, req *connect.Request[ssoreadyv1.AppUpdateOrganizationRequest]) (*connect.Response[ssoreadyv1.Organization], error) {
	res, err := s.Store.AppUpdateOrganization(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if err := appanalytics.Track(ctx, "Organization Updated", analytics.Properties{
		"environment_id":  res.EnvironmentId,
		"organization_id": res.Id,
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppDeleteOrganization(ctx context.Context, req *connect.Request[ssoreadyv1.AppDeleteOrganizationRequest]) (*connect.Response[emptypb.Empty], error) {
	res, err := s.Store.AppDeleteOrganization(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if err := appanalytics.Track(ctx, "Organization Deleted", analytics.Properties{
		"organization_id": req.Msg.OrganizationId,
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
