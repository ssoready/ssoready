package store

import (
	"context"

	"github.com/ssoready/ssoready/internal/store/queries"
)

type GetAPIKeyBySecretValueRequest struct {
	SecretValue string
}

func (s *Store) GetAPIKeyBySecretValue(ctx context.Context, req *GetAPIKeyBySecretValueRequest) (*queries.ApiKey, error) {
	apiKey, err := s.q.GetAPIKeyBySecretValue(ctx, req.SecretValue)
	if err != nil {
		return nil, err
	}

	return &apiKey, nil
}
