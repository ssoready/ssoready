package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/appauth"
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
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                startID,
		Limit:             int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var envs []*ssoreadyv1.Environment
	for _, qEnv := range qEnvs {
		envs = append(envs, &ssoreadyv1.Environment{
			Id:          idformat.Environment.Format(qEnv.ID),
			DisplayName: derefOrEmpty(qEnv.DisplayName),
			RedirectUrl: derefOrEmpty(qEnv.RedirectUrl),
		})
	}

	var nextPageToken string
	if len(envs) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(envs[limit].Id)
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
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                id,
	})
	if err != nil {
		return nil, err
	}

	return &ssoreadyv1.Environment{
		Id:          idformat.Environment.Format(qEnv.ID),
		DisplayName: derefOrEmpty(qEnv.DisplayName),
		RedirectUrl: derefOrEmpty(qEnv.RedirectUrl),
	}, nil
}
