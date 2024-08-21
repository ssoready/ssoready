package store

import (
	"context"

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
		CustomAuthDomain: derefOrEmpty(qEnv.CustomAuthDomain),
	}, nil
}

func (s *Store) UpdateEnvironmentCustomDomainSettings(ctx context.Context, req *ssoreadyv1.UpdateEnvironmentCustomDomainSettingsRequest) (*ssoreadyv1.UpdateEnvironmentCustomDomainSettingsResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

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

	if _, err := q.UpdateEnvironmentCustomAuthDomain(ctx, queries.UpdateEnvironmentCustomAuthDomainParams{
		ID:               id,
		CustomAuthDomain: &req.CustomAuthDomain,
	}); err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &ssoreadyv1.UpdateEnvironmentCustomDomainSettingsResponse{}, nil
}
