package apiservice

import (
	"context"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) ListSCIMDirectories(ctx context.Context, req *connect.Request[ssoreadyv1.ListSCIMDirectoriesRequest]) (*connect.Response[ssoreadyv1.ListSCIMDirectoriesResponse], error) {
	res, err := s.Store.ListSCIMDirectories(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) GetSCIMDirectory(ctx context.Context, req *connect.Request[ssoreadyv1.GetSCIMDirectoryRequest]) (*connect.Response[ssoreadyv1.GetSCIMDirectoryResponse], error) {
	res, err := s.Store.GetSCIMDirectory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) CreateSCIMDirectory(ctx context.Context, req *connect.Request[ssoreadyv1.CreateSCIMDirectoryRequest]) (*connect.Response[ssoreadyv1.CreateSCIMDirectoryResponse], error) {
	res, err := s.Store.CreateSCIMDirectory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) UpdateSCIMDirectory(ctx context.Context, req *connect.Request[ssoreadyv1.UpdateSCIMDirectoryRequest]) (*connect.Response[ssoreadyv1.UpdateSCIMDirectoryResponse], error) {
	res, err := s.Store.UpdateSCIMDirectory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) RotateSCIMDirectoryBearerToken(ctx context.Context, req *connect.Request[ssoreadyv1.RotateSCIMDirectoryBearerTokenRequest]) (*connect.Response[ssoreadyv1.RotateSCIMDirectoryBearerTokenResponse], error) {
	res, err := s.Store.RotateSCIMDirectoryBearerToken(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
