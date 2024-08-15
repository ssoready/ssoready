package store

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) ListEnvironments(ctx context.Context, req *ssoreadyv1.ListEnvironmentsRequest) (*ssoreadyv1.ListEnvironmentsResponse, error) {
	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, err
	}

	limit := 10
	qEnvs, err := s.q.ListEnvironments(ctx, queries.ListEnvironmentsParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                startID,
		Limit:             int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var envs []*ssoreadyv1.Environment
	for _, qEnv := range qEnvs {
		envs = append(envs, parseEnvironment(qEnv))
	}

	var nextPageToken string
	if len(envs) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(qEnvs[limit].ID)
		envs = envs[:limit]
	}

	return &ssoreadyv1.ListEnvironmentsResponse{
		Environments:  envs,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Store) GetEnvironment(ctx context.Context, req *ssoreadyv1.GetEnvironmentRequest) (*ssoreadyv1.Environment, error) {
	id, err := idformat.Environment.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	qEnv, err := s.q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
	})
	if err != nil {
		return nil, err
	}

	return parseEnvironment(qEnv), nil
}

func (s *Store) CreateEnvironment(ctx context.Context, req *ssoreadyv1.CreateEnvironmentRequest) (*ssoreadyv1.Environment, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	// attempt to parse redirect URL to ensure it's a URL
	if _, err := url.Parse(req.Environment.RedirectUrl); err != nil {
		return nil, fmt.Errorf("invalid redirect url: %w", err)
	}

	var authURL *string
	if req.Environment.AuthUrl != "" {
		authURL = &req.Environment.AuthUrl
	}

	qEnv, err := q.CreateEnvironment(ctx, queries.CreateEnvironmentParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                uuid.New(),
		RedirectUrl:       &req.Environment.RedirectUrl,
		DisplayName:       &req.Environment.DisplayName,
		AuthUrl:           authURL,
	})
	if err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return parseEnvironment(qEnv), nil
}

func (s *Store) UpdateEnvironment(ctx context.Context, req *ssoreadyv1.UpdateEnvironmentRequest) (*ssoreadyv1.Environment, error) {
	id, err := idformat.Environment.Parse(req.Environment.Id)
	if err != nil {
		return nil, err
	}

	// attempt to parse redirect URL to ensure it's a URL
	if _, err := url.Parse(req.Environment.RedirectUrl); err != nil {
		return nil, fmt.Errorf("invalid redirect url: %w", err)
	}

	// attempt to parse oauth redirect URL to ensure it's a URL
	if _, err := url.Parse(req.Environment.OauthRedirectUri); err != nil {
		return nil, fmt.Errorf("invalid redirect url: %w", err)
	}

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	// authz check
	if _, err := q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
	}); err != nil {
		return nil, err
	}

	var authURL *string
	if req.Environment.AuthUrl != "" {
		authURL = &req.Environment.AuthUrl
	}

	qEnv, err := q.UpdateEnvironment(ctx, queries.UpdateEnvironmentParams{
		ID:               id,
		DisplayName:      &req.Environment.DisplayName,
		RedirectUrl:      &req.Environment.RedirectUrl,
		AuthUrl:          authURL,
		OauthRedirectUri: &req.Environment.OauthRedirectUri,
	})
	if err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return parseEnvironment(qEnv), nil
}

func parseEnvironment(qEnv queries.Environment) *ssoreadyv1.Environment {
	return &ssoreadyv1.Environment{
		Id:               idformat.Environment.Format(qEnv.ID),
		DisplayName:      derefOrEmpty(qEnv.DisplayName),
		RedirectUrl:      derefOrEmpty(qEnv.RedirectUrl),
		AuthUrl:          derefOrEmpty(qEnv.AuthUrl),
		OauthRedirectUri: derefOrEmpty(qEnv.OauthRedirectUri),
	}
}
