package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) CreateAdminSetupURL(ctx context.Context, req *connect.Request[ssoreadyv1.CreateAdminSetupURLRequest]) (*connect.Response[ssoreadyv1.CreateAdminSetupURLResponse], error) {
	res, err := s.Store.CreateAdminSetupURL(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}
