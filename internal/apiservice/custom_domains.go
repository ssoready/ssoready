package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/cloudflare/cloudflare-go"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Service) GetEnvironmentCustomDomainSettings(ctx context.Context, req *connect.Request[ssoreadyv1.GetEnvironmentCustomDomainSettingsRequest]) (*connect.Response[ssoreadyv1.GetEnvironmentCustomDomainSettingsResponse], error) {
	res, err := s.Store.GetEnvironmentCustomDomainSettings(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if res.CustomAuthDomain != "" {
		hostname, err := s.getCloudflareCustomHostname(ctx, s.CustomAuthDomainCloudflareZoneID, res.CustomAuthDomain)
		if err != nil {
			return nil, fmt.Errorf("cloudflare: get custom auth hostname: %w", err)
		}

		if hostname != nil {
			res.CustomAuthDomainConfigured = hostname.Status == cloudflare.ACTIVE
		}
	}

	if res.CustomAdminDomain != "" {
		hostname, err := s.getCloudflareCustomHostname(ctx, s.CustomAdminDomainCloudflareZoneID, res.CustomAdminDomain)
		if err != nil {
			return nil, fmt.Errorf("cloudflare: get custom admin hostname: %w", err)
		}

		if hostname != nil {
			res.CustomAdminDomainConfigured = hostname.Status == cloudflare.ACTIVE
		}
	}

	// promoting a domain is idempotent
	if res.CustomAuthDomainConfigured {
		if err := s.Store.PromoteEnvironmentCustomAuthDomain(ctx, req.Msg.EnvironmentId); err != nil {
			return nil, err
		}
	}

	if res.CustomAdminDomainConfigured {
		if err := s.Store.PromoteEnvironmentCustomAdminDomain(ctx, req.Msg.EnvironmentId); err != nil {
			return nil, err
		}
	}

	res.CustomAuthDomainCnameValue = s.CustomAuthDomainCloudflareCNAMEValue
	res.CustomAdminDomainCnameValue = s.CustomAdminDomainCloudflareCNAMEValue
	return connect.NewResponse(res), nil
}

func (s *Service) UpdateEnvironmentCustomDomainSettings(ctx context.Context, req *connect.Request[ssoreadyv1.UpdateEnvironmentCustomDomainSettingsRequest]) (*connect.Response[ssoreadyv1.UpdateEnvironmentCustomDomainSettingsResponse], error) {
	appOrg, err := s.Store.GetAppOrganization(ctx, &ssoreadyv1.GetAppOrganizationRequest{})
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if !appOrg.EntitledCustomDomains {
		return nil, fmt.Errorf("not entitled to custom domains")
	}

	if req.Msg.CustomAuthDomain != "" {
		if _, err := s.CloudflareClient.CreateCustomHostname(ctx, s.CustomAuthDomainCloudflareZoneID, cloudflare.CustomHostname{
			Hostname: req.Msg.CustomAuthDomain,
			SSL: &cloudflare.CustomHostnameSSL{
				Method: "http",
				Type:   "dv",
			},
		}); err != nil {
			return nil, fmt.Errorf("cloudflare: create auth custom hostname: %w", err)
		}
	}

	if req.Msg.CustomAdminDomain != "" {
		if _, err := s.CloudflareClient.CreateCustomHostname(ctx, s.CustomAdminDomainCloudflareZoneID, cloudflare.CustomHostname{
			Hostname: req.Msg.CustomAdminDomain,
			SSL: &cloudflare.CustomHostnameSSL{
				Method: "http",
				Type:   "dv",
			},
		}); err != nil {
			return nil, fmt.Errorf("cloudflare: create admin custom hostname: %w", err)
		}
	}

	res, err := s.Store.UpdateEnvironmentCustomDomainSettings(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}

func (s *Service) CheckEnvironmentCustomDomainSettingsCertificates(ctx context.Context, req *connect.Request[ssoreadyv1.CheckEnvironmentCustomDomainSettingsCertificatesRequest]) (*connect.Response[ssoreadyv1.CheckEnvironmentCustomDomainSettingsCertificatesResponse], error) {
	// TODO: this endpoint is redundant with Cloudflare, because they auto-check
	// TLS settings where fly.io does not (the initial motivation for this
	// endpoint).
	//
	// That said, this endpoint is preserved because CloudFlare does let you
	// tighten their periodic DNS checks:
	//
	// https://developers.cloudflare.com/cloudflare-for-platforms/cloudflare-for-saas/domain-support/hostname-validation/realtime-validation/#use-when
	//
	// This functionality is not implemented here, but may be added later if the
	// default Cloudflare DNS check schedule isn't acceptable.

	env, err := s.Store.GetEnvironmentCustomDomainSettings(ctx, &ssoreadyv1.GetEnvironmentCustomDomainSettingsRequest{
		EnvironmentId: req.Msg.EnvironmentId,
	})
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	var customAuthDomainConfigured bool
	if env.CustomAuthDomain != "" {
		hostname, err := s.getCloudflareCustomHostname(ctx, s.CustomAuthDomainCloudflareZoneID, env.CustomAuthDomain)
		if err != nil {
			return nil, fmt.Errorf("cloudflare: get custom auth hostname: %w", err)
		}

		if hostname != nil {
			customAuthDomainConfigured = hostname.Status == cloudflare.ACTIVE
		}
	}

	var customAdminDomainConfigured bool
	if env.CustomAdminDomain != "" {
		hostname, err := s.getCloudflareCustomHostname(ctx, s.CustomAdminDomainCloudflareZoneID, env.CustomAdminDomain)
		if err != nil {
			return nil, fmt.Errorf("cloudflare: get custom admin hostname: %w", err)
		}

		if hostname != nil {
			customAdminDomainConfigured = hostname.Status == cloudflare.ACTIVE
		}
	}

	if customAuthDomainConfigured {
		if err := s.Store.PromoteEnvironmentCustomAuthDomain(ctx, req.Msg.EnvironmentId); err != nil {
			return nil, err
		}
	}

	if customAdminDomainConfigured {
		if err := s.Store.PromoteEnvironmentCustomAdminDomain(ctx, req.Msg.EnvironmentId); err != nil {
			return nil, err
		}
	}

	return connect.NewResponse(&ssoreadyv1.CheckEnvironmentCustomDomainSettingsCertificatesResponse{
		CustomAuthDomainConfigured:  customAuthDomainConfigured,
		CustomAdminDomainConfigured: customAdminDomainConfigured,
	}), nil
}

func (s *Service) getCloudflareCustomHostname(ctx context.Context, zoneID, hostname string) (*cloudflare.CustomHostname, error) {
	customHostnames, _, err := s.CloudflareClient.CustomHostnames(ctx, zoneID, 1, cloudflare.CustomHostname{
		Hostname: hostname,
	})
	if err != nil {
		return nil, fmt.Errorf("cloudflare: list custom hostnames: %w", err)
	}

	if len(customHostnames) == 0 {
		return nil, nil
	}
	return &customHostnames[0], nil
}
