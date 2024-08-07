package apiservice

import (
	"context"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) AppListSCIMUsers(ctx context.Context, req *connect.Request[ssoreadyv1.AppListSCIMUsersRequest]) (*connect.Response[ssoreadyv1.AppListSCIMUsersResponse], error) {
	res, err := s.Store.AppListSCIMUsers(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppGetSCIMUser(ctx context.Context, req *connect.Request[ssoreadyv1.AppGetSCIMUserRequest]) (*connect.Response[ssoreadyv1.SCIMUser], error) {
	res, err := s.Store.AppGetSCIMUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
