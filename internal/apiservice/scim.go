package apiservice

import (
	"context"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) ListSCIMUsers(ctx context.Context, req *connect.Request[ssoreadyv1.ListSCIMUsersRequest]) (*connect.Response[ssoreadyv1.ListSCIMUsersResponse], error) {
	res, err := s.Store.ListSCIMUsers(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) GetSCIMUser(ctx context.Context, req *connect.Request[ssoreadyv1.GetSCIMUserRequest]) (*connect.Response[ssoreadyv1.GetSCIMUserResponse], error) {
	res, err := s.Store.GetSCIMUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) ListSCIMGroups(ctx context.Context, req *connect.Request[ssoreadyv1.ListSCIMGroupsRequest]) (*connect.Response[ssoreadyv1.ListSCIMGroupsResponse], error) {
	panic("todo")
}

func (s *Service) GetSCIMGroup(ctx context.Context, req *connect.Request[ssoreadyv1.GetSCIMGroupRequest]) (*connect.Response[ssoreadyv1.GetSCIMGroupResponse], error) {
	panic("todo")
}
