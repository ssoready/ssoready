package store

import (
	"context"
	"crypto/sha256"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Store) ListSAMLOAuthClients(ctx context.Context, req *ssoreadyv1.ListSAMLOAuthClientsRequest) (*ssoreadyv1.ListSAMLOAuthClientsResponse, error) {
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
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                envID,
	}); err != nil {
		return nil, err
	}

	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, err
	}

	limit := 10
	qSAMLOAuthClients, err := q.ListSAMLOAuthClients(ctx, queries.ListSAMLOAuthClientsParams{
		EnvironmentID: envID,
		ID:            startID,
		Limit:         int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var samlOAuthClients []*ssoreadyv1.SAMLOAuthClient
	for _, qSAMLOAuthClient := range qSAMLOAuthClients {
		samlOAuthClients = append(samlOAuthClients, parseSAMLOAuthClient(qSAMLOAuthClient))
	}

	var nextPageToken string
	if len(samlOAuthClients) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(samlOAuthClients[limit].Id)
		samlOAuthClients = samlOAuthClients[:limit]
	}

	return &ssoreadyv1.ListSAMLOAuthClientsResponse{
		SamlOauthClients: samlOAuthClients,
		NextPageToken:    nextPageToken,
	}, nil
}

func (s *Store) GetSAMLOAuthClient(ctx context.Context, req *ssoreadyv1.GetSAMLOAuthClientRequest) (*ssoreadyv1.SAMLOAuthClient, error) {
	id, err := idformat.SAMLOAuthClient.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	// idor check
	qSAMLOAuthClient, err := q.GetSAMLOAuthClient(ctx, queries.GetSAMLOAuthClientParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
	})
	if err != nil {
		return nil, err
	}

	return parseSAMLOAuthClient(qSAMLOAuthClient), nil
}

func (s *Store) CreateSAMLOAuthClient(ctx context.Context, req *ssoreadyv1.CreateSAMLOAuthClientRequest) (*ssoreadyv1.SAMLOAuthClient, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	envID, err := idformat.Environment.Parse(req.SamlOauthClient.EnvironmentId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                envID,
	}); err != nil {
		return nil, err
	}

	secretValue := uuid.New()
	secretValueSHA := sha256.Sum256(secretValue[:])

	qSAMLOAuthClient, err := q.CreateSAMLOAuthClient(ctx, queries.CreateSAMLOAuthClientParams{
		EnvironmentID:      envID,
		ID:                 uuid.New(),
		ClientSecretSha256: secretValueSHA[:],
	})
	if err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	SAMLOAuthClient := parseSAMLOAuthClient(qSAMLOAuthClient)
	SAMLOAuthClient.ClientSecret = idformat.SAMLOAuthClientSecret.Format(secretValue)
	return SAMLOAuthClient, nil
}

func (s *Store) DeleteSAMLOAuthClient(ctx context.Context, req *ssoreadyv1.DeleteSAMLOAuthClientRequest) (*emptypb.Empty, error) {
	id, err := idformat.SAMLOAuthClient.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	// authz check
	if _, err := q.GetSAMLOAuthClient(ctx, queries.GetSAMLOAuthClientParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
	}); err != nil {
		return nil, err
	}

	if err := q.DeleteSAMLOAuthClient(ctx, id); err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func parseSAMLOAuthClient(qSAMLOAuthClient queries.SamlOauthClient) *ssoreadyv1.SAMLOAuthClient {
	return &ssoreadyv1.SAMLOAuthClient{
		Id:            idformat.SAMLOAuthClient.Format(qSAMLOAuthClient.ID),
		EnvironmentId: idformat.Environment.Format(qSAMLOAuthClient.EnvironmentID),
		ClientSecret:  "", // intentionally left blank
	}
}
