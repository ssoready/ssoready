package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) GetEnvironmentCustomDomainSettings(ctx context.Context, req *connect.Request[ssoreadyv1.GetEnvironmentCustomDomainSettingsRequest]) (*connect.Response[ssoreadyv1.GetEnvironmentCustomDomainSettingsResponse], error) {
	res, err := s.Store.GetEnvironmentCustomDomainSettings(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}
