package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/ssoready/ssoready/internal/flyio"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) GetEnvironmentCustomDomainSettings(ctx context.Context, req *connect.Request[ssoreadyv1.GetEnvironmentCustomDomainSettingsRequest]) (*connect.Response[ssoreadyv1.GetEnvironmentCustomDomainSettingsResponse], error) {
	res, err := s.Store.GetEnvironmentCustomDomainSettings(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if res.CustomAuthDomain != "" {
		certRes, err := s.FlyioClient.GetCertificate(ctx, &flyio.GetCertificateRequest{
			AppID:    s.FlyioAuthProxyAppID,
			Hostname: res.CustomAuthDomain,
		})
		if err != nil {
			return nil, fmt.Errorf("flyio: %w", err)
		}

		res.CustomAuthDomainConfigured = certRes.Certificate.Configured
	}

	return connect.NewResponse(res), nil
}
