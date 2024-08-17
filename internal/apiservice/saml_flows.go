package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) AppListSAMLFlows(ctx context.Context, req *connect.Request[ssoreadyv1.AppListSAMLFlowsRequest]) (*connect.Response[ssoreadyv1.AppListSAMLFlowsResponse], error) {
	res, err := s.Store.AppListSAMLFlows(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppGetSAMLFlow(ctx context.Context, req *connect.Request[ssoreadyv1.AppGetSAMLFlowRequest]) (*connect.Response[ssoreadyv1.SAMLFlow], error) {
	res, err := s.Store.AppGetSAMLFlow(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}
