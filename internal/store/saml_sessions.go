package store

import (
	"context"

	"github.com/ssoready/ssoready/internal/appauth"
	"github.com/ssoready/ssoready/internal/store/queries"
)

type GetSAMLSessionBySecretAccessTokenRequest struct {
	SecretAccessToken string
}

func (s *Store) GetSAMLSessionBySecretAccessToken(ctx context.Context, req *GetSAMLSessionBySecretAccessTokenRequest) (*queries.SamlSession, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	samlSession, err := q.GetSAMLSessionBySecretAccessToken(ctx, queries.GetSAMLSessionBySecretAccessTokenParams{
		AppOrganizationID: appauth.OrgID(ctx),
		SecretAccessToken: &req.SecretAccessToken,
	})
	if err != nil {
		return nil, err
	}

	return &samlSession, nil
}
