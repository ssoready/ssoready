package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/store/idformat"
)

type GetAPIKeyBySecretTokenRequest struct {
	Token string
}

type GetAPIKeyBySecretTokenResponse struct {
	ID                string
	AppOrganizationID uuid.UUID
}

func (s *Store) GetAPIKeyBySecretToken(ctx context.Context, req *GetAPIKeyBySecretTokenRequest) (*GetAPIKeyBySecretTokenResponse, error) {
	apiKey, err := s.q.GetAPIKeyBySecretValue(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	return &GetAPIKeyBySecretTokenResponse{
		ID:                idformat.APIKey.Format(apiKey.ID),
		AppOrganizationID: apiKey.AppOrganizationID,
	}, nil
}
