package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) ListSAMLFlows(ctx context.Context, req *connect.Request[ssoreadyv1.ListSAMLFlowsRequest]) (*connect.Response[ssoreadyv1.ListSAMLFlowsResponse], error) {
	res, err := s.Store.ListSAMLFlows(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) GetSAMLFlow(ctx context.Context, req *connect.Request[ssoreadyv1.GetSAMLFlowRequest]) (*connect.Response[ssoreadyv1.SAMLFlow], error) {
	res, err := s.Store.GetSAMLFlow(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}
