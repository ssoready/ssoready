package store

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

type AuthOAuthGetClientRequest struct {
	SAMLOAuthClientID string
}

type AuthOAuthGetClientResponse struct {
	AppOrgID          uuid.UUID
	EnvID             string
	SAMLOAuthClientID string
}

func (s *Store) AuthOAuthGetClient(ctx context.Context, req *AuthOAuthGetClientRequest) (*AuthOAuthGetClientResponse, error) {
	clientID, err := idformat.SAMLOAuthClient.Parse(req.SAMLOAuthClientID)
	if err != nil {
		return nil, err
	}

	client, err := s.q.AuthGetSAMLOAuthClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	return &AuthOAuthGetClientResponse{
		AppOrgID:          client.AppOrganizationID,
		EnvID:             idformat.Environment.Format(client.EnvironmentID),
		SAMLOAuthClientID: idformat.SAMLOAuthClient.Format(client.ID),
	}, nil
}

type AuthOAuthGetClientWithSecretRequest struct {
	SAMLOAuthClientID     string
	SAMLOAuthClientSecret string
}

type AuthOAuthGetClientWithSecretResponse struct {
	AppOrgID          uuid.UUID
	EnvID             string
	SAMLOAuthClientID string
}

func (s *Store) AuthOAuthGetClientWithSecret(ctx context.Context, req *AuthOAuthGetClientWithSecretRequest) (*AuthOAuthGetClientWithSecretResponse, error) {
	clientID, err := idformat.SAMLOAuthClient.Parse(req.SAMLOAuthClientID)
	if err != nil {
		return nil, fmt.Errorf("parse client id: %w", err)
	}

	clientSecret, err := idformat.SAMLOAuthClientSecret.Parse(req.SAMLOAuthClientSecret)
	if err != nil {
		return nil, fmt.Errorf("parse client secret: %w", err)
	}

	clientSecretSHA := sha256.Sum256(clientSecret[:])

	client, err := s.q.AuthGetSAMLOAuthClientWithSecret(ctx, queries.AuthGetSAMLOAuthClientWithSecretParams{
		ID:                 clientID,
		ClientSecretSha256: clientSecretSHA[:],
	})
	if err != nil {
		return nil, err
	}

	return &AuthOAuthGetClientWithSecretResponse{
		AppOrgID:          client.AppOrganizationID,
		EnvID:             idformat.Environment.Format(client.EnvironmentID),
		SAMLOAuthClientID: idformat.SAMLOAuthClient.Format(client.ID),
	}, nil
}
