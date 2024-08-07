package apiservice

import (
	"context"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) AppListSCIMGroups(ctx context.Context, req *connect.Request[ssoreadyv1.AppListSCIMGroupsRequest]) (*connect.Response[ssoreadyv1.AppListSCIMGroupsResponse], error) {
	res, err := s.Store.AppListSCIMGroups(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppGetSCIMGroup(ctx context.Context, req *connect.Request[ssoreadyv1.AppGetSCIMGroupRequest]) (*connect.Response[ssoreadyv1.SCIMGroup], error) {
	res, err := s.Store.AppGetSCIMGroup(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
