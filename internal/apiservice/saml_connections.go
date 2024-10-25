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

func (s *Service) AppListSAMLConnections(ctx context.Context, req *connect.Request[ssoreadyv1.AppListSAMLConnectionsRequest]) (*connect.Response[ssoreadyv1.AppListSAMLConnectionsResponse], error) {
	res, err := s.Store.AppListSAMLConnections(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppGetSAMLConnection(ctx context.Context, req *connect.Request[ssoreadyv1.AppGetSAMLConnectionRequest]) (*connect.Response[ssoreadyv1.SAMLConnection], error) {
	res, err := s.Store.AppGetSAMLConnection(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppCreateSAMLConnection(ctx context.Context, req *connect.Request[ssoreadyv1.AppCreateSAMLConnectionRequest]) (*connect.Response[ssoreadyv1.SAMLConnection], error) {
	res, err := s.Store.AppCreateSAMLConnection(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if err := appanalytics.Track(ctx, "SAML Connection Created", analytics.Properties{
		"organization_id":    res.OrganizationId,
		"saml_connection_id": res.Id,
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppUpdateSAMLConnection(ctx context.Context, req *connect.Request[ssoreadyv1.AppUpdateSAMLConnectionRequest]) (*connect.Response[ssoreadyv1.SAMLConnection], error) {
	res, err := s.Store.AppUpdateSAMLConnection(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if err := appanalytics.Track(ctx, "SAML Connection Updated", analytics.Properties{
		"organization_id":    res.OrganizationId,
		"saml_connection_id": res.Id,
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppDeleteSAMLConnection(ctx context.Context, req *connect.Request[ssoreadyv1.AppDeleteSAMLConnectionRequest]) (*connect.Response[emptypb.Empty], error) {
	res, err := s.Store.AppDeleteSAMLConnection(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if err := appanalytics.Track(ctx, "SAML Connection Deleted", analytics.Properties{
		"saml_connection_id": req.Msg.SamlConnectionId,
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
