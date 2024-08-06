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

func (s *Service) GetSCIMDirectory(ctx context.Context, req *connect.Request[ssoreadyv1.GetSCIMDirectoryRequest]) (*connect.Response[ssoreadyv1.SCIMDirectory], error) {
	res, err := s.Store.GetSCIMDirectory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) CreateSCIMDirectory(ctx context.Context, req *connect.Request[ssoreadyv1.CreateSCIMDirectoryRequest]) (*connect.Response[ssoreadyv1.SCIMDirectory], error) {
	res, err := s.Store.CreateSCIMDirectory(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
