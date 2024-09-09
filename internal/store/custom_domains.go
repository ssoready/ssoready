package store

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) GetEnvironmentCustomDomainSettings(ctx context.Context, req *ssoreadyv1.GetEnvironmentCustomDomainSettingsRequest) (*ssoreadyv1.GetEnvironmentCustomDomainSettingsResponse, error) {
	id, err := idformat.Environment.Parse(req.EnvironmentId)
	if err != nil {
		return nil, err
	}

	// also acts as an authz check
	qEnv, err := s.q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
	})
	if err != nil {
		return nil, err
	}

	return &ssoreadyv1.GetEnvironmentCustomDomainSettingsResponse{
		CustomAuthDomain:  derefOrEmpty(qEnv.CustomAuthDomain),
		CustomAdminDomain: derefOrEmpty(qEnv.CustomAdminDomain),
	}, nil
}

func (s *Store) UpdateEnvironmentCustomDomainSettings(ctx context.Context, req *ssoreadyv1.UpdateEnvironmentCustomDomainSettingsRequest) (*ssoreadyv1.UpdateEnvironmentCustomDomainSettingsResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	// entitlement check
	qAppOrg, err := q.GetAppOrganizationByID(ctx, authn.AppOrgID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get app org by id: %w", err)
	}

	if !derefOrEmpty(qAppOrg.EntitledCustomDomains) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("app organization is not entitled to custom domains"))
	}

	id, err := idformat.Environment.Parse(req.EnvironmentId)
	if err != nil {
		return nil, err
	}

	if _, err := q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
	}); err != nil {
		return nil, err
	}

	if req.CustomAuthDomain != "" {
		if _, err := q.UpdateEnvironmentCustomAuthDomain(ctx, queries.UpdateEnvironmentCustomAuthDomainParams{
			ID:               id,
			CustomAuthDomain: &req.CustomAuthDomain,
		}); err != nil {
			return nil, err
		}
	}

	if req.CustomAdminDomain != "" {
		if _, err := q.UpdateEnvironmentCustomAdminDomain(ctx, queries.UpdateEnvironmentCustomAdminDomainParams{
			ID:                id,
			CustomAdminDomain: &req.CustomAdminDomain,
		}); err != nil {
			return nil, err
		}
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &ssoreadyv1.UpdateEnvironmentCustomDomainSettingsResponse{}, nil
}

func (s *Store) PromoteEnvironmentCustomAuthDomain(ctx context.Context, environmentID string) error {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return err
	}
	defer rollback()

	id, err := idformat.Environment.Parse(environmentID)
	if err != nil {
		return err
	}

	qEnv, err := q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
	})
	if err != nil {
		return err
	}

	authURL := fmt.Sprintf("https://%s", *qEnv.CustomAuthDomain)
	if _, err := q.UpdateEnvironmentAuthURL(ctx, queries.UpdateEnvironmentAuthURLParams{
		ID:      id,
		AuthUrl: &authURL,
	}); err != nil {
		return err
	}

	if err := commit(); err != nil {
		return err
	}

	return nil
}

func (s *Store) PromoteEnvironmentCustomAdminDomain(ctx context.Context, environmentID string) error {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return err
	}
	defer rollback()

	id, err := idformat.Environment.Parse(environmentID)
	if err != nil {
		return err
	}

	qEnv, err := q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
	})
	if err != nil {
		return err
	}

	adminURL := fmt.Sprintf("https://%s", *qEnv.CustomAdminDomain)
	if _, err := q.UpdateEnvironmentAdminURL(ctx, queries.UpdateEnvironmentAdminURLParams{
		ID:       id,
		AdminUrl: &adminURL,
	}); err != nil {
		return err
	}

	if err := commit(); err != nil {
		return err
	}

	return nil
}
