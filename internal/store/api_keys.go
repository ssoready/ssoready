package store

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/appauth"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GetAPIKeyBySecretTokenRequest struct {
	Token string
}

type GetAPIKeyBySecretTokenResponse struct {
	ID                string
	AppOrganizationID uuid.UUID
	EnvironmentID     uuid.UUID
}

func (s *Store) GetAPIKeyBySecretToken(ctx context.Context, req *GetAPIKeyBySecretTokenRequest) (*GetAPIKeyBySecretTokenResponse, error) {
	secretValue, err := idformat.APISecretKey.Parse(req.Token)
	if err != nil {
		return nil, fmt.Errorf("parse api secret key: %w", err)
	}

	secretValueSHA := sha256.Sum256(secretValue[:])

	apiKey, err := s.q.GetAPIKeyBySecretValueSHA256(ctx, secretValueSHA[:])
	if err != nil {
		return nil, fmt.Errorf("get api key by secret value: %w", err)
	}

	return &GetAPIKeyBySecretTokenResponse{
		ID:                idformat.APIKey.Format(apiKey.ID),
		AppOrganizationID: apiKey.AppOrganizationID,
		EnvironmentID:     apiKey.EnvironmentID,
	}, nil
}

func (s *Store) ListAPIKeys(ctx context.Context, req *ssoreadyv1.ListAPIKeysRequest) (*ssoreadyv1.ListAPIKeysResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	envID, err := idformat.Environment.Parse(req.EnvironmentId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                envID,
	}); err != nil {
		return nil, err
	}

	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, err
	}

	limit := 10
	qAPIKeys, err := q.ListAPIKeys(ctx, queries.ListAPIKeysParams{
		EnvironmentID: envID,
		ID:            startID,
		Limit:         int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var apiKeys []*ssoreadyv1.APIKey
	for _, qAPIKey := range qAPIKeys {
		apiKeys = append(apiKeys, parseAPIKey(qAPIKey))
	}

	var nextPageToken string
	if len(apiKeys) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(apiKeys[limit].Id)
		apiKeys = apiKeys[:limit]
	}

	return &ssoreadyv1.ListAPIKeysResponse{
		ApiKeys:       apiKeys,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Store) GetAPIKey(ctx context.Context, req *ssoreadyv1.GetAPIKeyRequest) (*ssoreadyv1.APIKey, error) {
	id, err := idformat.APIKey.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	// idor check
	qAPIKey, err := q.GetAPIKey(ctx, queries.GetAPIKeyParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                id,
	})
	if err != nil {
		return nil, err
	}

	return parseAPIKey(qAPIKey), nil
}

func (s *Store) CreateAPIKey(ctx context.Context, req *ssoreadyv1.CreateAPIKeyRequest) (*ssoreadyv1.APIKey, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	envID, err := idformat.Environment.Parse(req.ApiKey.EnvironmentId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                envID,
	}); err != nil {
		return nil, err
	}

	secretValue := uuid.New()
	secretValueSHA := sha256.Sum256(secretValue[:])

	qAPIKey, err := q.CreateAPIKey(ctx, queries.CreateAPIKeyParams{
		EnvironmentID:     envID,
		ID:                uuid.New(),
		SecretValueSha256: secretValueSHA[:],
	})
	if err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	apiKey := parseAPIKey(qAPIKey)
	apiKey.SecretToken = idformat.APISecretKey.Format(secretValue)
	return apiKey, nil
}

func (s *Store) DeleteAPIKey(ctx context.Context, req *ssoreadyv1.DeleteAPIKeyRequest) (*emptypb.Empty, error) {
	id, err := idformat.APIKey.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	// authz check
	if _, err := q.GetAPIKey(ctx, queries.GetAPIKeyParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                id,
	}); err != nil {
		return nil, err
	}

	if err := q.DeleteAPIKey(ctx, id); err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func parseAPIKey(qAPIKey queries.ApiKey) *ssoreadyv1.APIKey {
	return &ssoreadyv1.APIKey{
		Id:            idformat.APIKey.Format(qAPIKey.ID),
		EnvironmentId: idformat.Environment.Format(qAPIKey.EnvironmentID),
		SecretToken:   "", // intentionally left blank
	}
}
