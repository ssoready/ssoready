package apiservice

import (
	"context"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) AppListSCIMDirectories(ctx context.Context, req *connect.Request[ssoreadyv1.AppListSCIMDirectoriesRequest]) (*connect.Response[ssoreadyv1.AppListSCIMDirectoriesResponse], error) {
	res, err := s.Store.AppListSCIMDirectories(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppGetSCIMDirectory(ctx context.Context, req *connect.Request[ssoreadyv1.AppGetSCIMDirectoryRequest]) (*connect.Response[ssoreadyv1.SCIMDirectory], error) {
	res, err := s.Store.AppGetSCIMDirectory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppCreateSCIMDirectory(ctx context.Context, req *connect.Request[ssoreadyv1.AppCreateSCIMDirectoryRequest]) (*connect.Response[ssoreadyv1.SCIMDirectory], error) {
	res, err := s.Store.AppCreateSCIMDirectory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppUpdateSCIMDirectory(ctx context.Context, req *connect.Request[ssoreadyv1.AppUpdateSCIMDirectoryRequest]) (*connect.Response[ssoreadyv1.SCIMDirectory], error) {
	res, err := s.Store.AppUpdateSCIMDirectory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppRotateSCIMDirectoryBearerToken(ctx context.Context, req *connect.Request[ssoreadyv1.AppRotateSCIMDirectoryBearerTokenRequest]) (*connect.Response[ssoreadyv1.AppRotateSCIMDirectoryBearerTokenResponse], error) {
	res, err := s.Store.AppRotateSCIMDirectoryBearerToken(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
