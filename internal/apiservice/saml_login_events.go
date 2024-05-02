package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) ListSAMLLoginEvents(ctx context.Context, req *connect.Request[ssoreadyv1.ListSAMLLoginEventsRequest]) (*connect.Response[ssoreadyv1.ListSAMLLoginEventsResponse], error) {
	res, err := s.Store.ListSAMLLoginEvents(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}
