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

	res.CustomAuthDomainCnameValue = s.FlyioAuthProxyAppCNAMEValue
	return connect.NewResponse(res), nil
}

func (s *Service) UpdateEnvironmentCustomDomainSettings(ctx context.Context, req *connect.Request[ssoreadyv1.UpdateEnvironmentCustomDomainSettingsRequest]) (*connect.Response[ssoreadyv1.UpdateEnvironmentCustomDomainSettingsResponse], error) {
	if req.Msg.CustomAuthDomain != "" {
		if _, err := s.FlyioClient.AddCertificate(ctx, &flyio.AddCertificateRequest{
			AppID:    s.FlyioAuthProxyAppID,
			Hostname: req.Msg.CustomAuthDomain,
		}); err != nil {
			return nil, fmt.Errorf("flyio: %w", err)
		}
	}

	res, err := s.Store.UpdateEnvironmentCustomDomainSettings(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}

func (s *Service) CheckEnvironmentCustomDomainSettingsCertificates(ctx context.Context, req *connect.Request[ssoreadyv1.CheckEnvironmentCustomDomainSettingsCertificatesRequest]) (*connect.Response[ssoreadyv1.CheckEnvironmentCustomDomainSettingsCertificatesResponse], error) {
	env, err := s.Store.GetEnvironmentCustomDomainSettings(ctx, &ssoreadyv1.GetEnvironmentCustomDomainSettingsRequest{
		EnvironmentId: req.Msg.EnvironmentId,
	})
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	checkRes, err := s.FlyioClient.CheckCertificate(ctx, &flyio.CheckCertificateRequest{
		AppID:    s.FlyioAuthProxyAppID,
		Hostname: env.CustomAuthDomain,
	})
	if err != nil {
		return nil, err
	}

	if checkRes.Certificate.Configured {
		if err := s.Store.PromoteEnvironmentCustomAuthDomain(ctx, req.Msg.EnvironmentId); err != nil {
			return nil, err
		}
	}

	return connect.NewResponse(&ssoreadyv1.CheckEnvironmentCustomDomainSettingsCertificatesResponse{
		CustomAuthDomainConfigured: checkRes.Certificate.Configured,
	}), nil
}
