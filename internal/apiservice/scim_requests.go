package apiservice

import (
	"context"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) AppListSCIMRequests(ctx context.Context, req *connect.Request[ssoreadyv1.AppListSCIMRequestsRequest]) (*connect.Response[ssoreadyv1.AppListSCIMRequestsResponse], error) {
	res, err := s.Store.AppListSCIMRequests(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppGetSCIMRequest(ctx context.Context, req *connect.Request[ssoreadyv1.AppGetSCIMRequestRequest]) (*connect.Response[ssoreadyv1.AppGetSCIMRequestResponse], error) {
	res, err := s.Store.AppGetSCIMRequest(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
