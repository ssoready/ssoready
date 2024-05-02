package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) ListSAMLLoginEventTimelineEntries(ctx context.Context, req *connect.Request[ssoreadyv1.ListSAMLLoginEventTimelineEntriesRequest]) (*connect.Response[ssoreadyv1.ListSAMLLoginEventTimelineEntriesResponse], error) {
	res, err := s.Store.ListSAMLLoginEventTimelineEntries(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}
