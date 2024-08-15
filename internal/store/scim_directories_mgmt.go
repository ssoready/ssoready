package store

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) ListSCIMDirectories(ctx context.Context, req *ssoreadyv1.ListSCIMDirectoriesRequest) (*ssoreadyv1.ListSCIMDirectoriesResponse, error) {
	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	orgID, err := idformat.Organization.Parse(req.OrganizationId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.ManagementGetOrganization(ctx, queries.ManagementGetOrganizationParams{
		EnvironmentID: envID,
		ID:            orgID,
	}); err != nil {
		return nil, err
	}

	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, err
	}

	limit := 10
	qSCIMDirectories, err := q.ListSCIMDirectories(ctx, queries.ListSCIMDirectoriesParams{
		OrganizationID: orgID,
		ID:             startID,
		Limit:          int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var scimDirectories []*ssoreadyv1.SCIMDirectory
	for _, qSCIMDirectory := range qSCIMDirectories {
		scimDirectories = append(scimDirectories, parseSCIMDirectory(qSCIMDirectory))
	}

	var nextPageToken string
	if len(scimDirectories) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(qSCIMDirectories[limit].ID)
		scimDirectories = scimDirectories[:limit]
	}

	return &ssoreadyv1.ListSCIMDirectoriesResponse{
		ScimDirectories: scimDirectories,
		NextPageToken:   nextPageToken,
	}, nil
}

func (s *Store) GetSCIMDirectory(ctx context.Context, req *ssoreadyv1.GetSCIMDirectoryRequest) (*ssoreadyv1.GetSCIMDirectoryResponse, error) {
	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	scimDirID, err := idformat.SCIMDirectory.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	qSCIMDir, err := q.ManagementGetSCIMDirectory(ctx, queries.ManagementGetSCIMDirectoryParams{
		EnvironmentID: envID,
		ID:            scimDirID,
	})
	if err != nil {
		return nil, fmt.Errorf("get scim directory: %w", err)
	}

	return &ssoreadyv1.GetSCIMDirectoryResponse{ScimDirectory: parseSCIMDirectory(qSCIMDir)}, nil
}

func (s *Store) CreateSCIMDirectory(ctx context.Context, req *ssoreadyv1.CreateSCIMDirectoryRequest) (*ssoreadyv1.CreateSCIMDirectoryResponse, error) {
	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	orgID, err := idformat.Organization.Parse(req.ScimDirectory.OrganizationId)
	if err != nil {
		return nil, fmt.Errorf("parse organization id: %w", err)
	}

	// idor check
	if _, err := q.ManagementGetOrganization(ctx, queries.ManagementGetOrganizationParams{
		EnvironmentID: envID,
		ID:            orgID,
	}); err != nil {
		return nil, err
	}

	env, err := q.GetEnvironmentByID(ctx, envID)
	if err != nil {
		return nil, err
	}

	authURL := s.defaultAuthURL
	if env.AuthUrl != nil {
		authURL = *env.AuthUrl
	}

	id := uuid.New()
	scimBaseURL := fmt.Sprintf("%s/v1/scim/%s", authURL, idformat.SCIMDirectory.Format(id))
	qSCIMDirectory, err := q.CreateSCIMDirectory(ctx, queries.CreateSCIMDirectoryParams{
		ID:             id,
		OrganizationID: orgID,
		IsPrimary:      req.ScimDirectory.Primary,
		ScimBaseUrl:    scimBaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create scim directory: %w", err)
	}

	if qSCIMDirectory.IsPrimary {
		if err := q.UpdatePrimarySCIMDirectory(ctx, queries.UpdatePrimarySCIMDirectoryParams{
			ID:             qSCIMDirectory.ID,
			OrganizationID: qSCIMDirectory.OrganizationID,
		}); err != nil {
			return nil, fmt.Errorf("update primary scim directory: %w", err)
		}
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ssoreadyv1.CreateSCIMDirectoryResponse{ScimDirectory: parseSCIMDirectory(qSCIMDirectory)}, nil
}

func (s *Store) UpdateSCIMDirectory(ctx context.Context, req *ssoreadyv1.UpdateSCIMDirectoryRequest) (*ssoreadyv1.UpdateSCIMDirectoryResponse, error) {
	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	scimDirID, err := idformat.SCIMDirectory.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	if _, err := q.ManagementGetSCIMDirectory(ctx, queries.ManagementGetSCIMDirectoryParams{
		EnvironmentID: envID,
		ID:            scimDirID,
	}); err != nil {
		return nil, fmt.Errorf("get scim directory: %w", err)
	}

	qSCIMDir, err := q.UpdateSCIMDirectory(ctx, queries.UpdateSCIMDirectoryParams{
		ID:        scimDirID,
		IsPrimary: req.ScimDirectory.Primary,
	})
	if err != nil {
		return nil, fmt.Errorf("update scim directory: %w", err)
	}

	if qSCIMDir.IsPrimary {
		if err := q.UpdatePrimarySCIMDirectory(ctx, queries.UpdatePrimarySCIMDirectoryParams{
			ID:             qSCIMDir.ID,
			OrganizationID: qSCIMDir.OrganizationID,
		}); err != nil {
			return nil, fmt.Errorf("update primary scim directory: %w", err)
		}
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ssoreadyv1.UpdateSCIMDirectoryResponse{ScimDirectory: parseSCIMDirectory(qSCIMDir)}, nil
}

func (s *Store) RotateSCIMDirectoryBearerToken(ctx context.Context, req *ssoreadyv1.RotateSCIMDirectoryBearerTokenRequest) (*ssoreadyv1.RotateSCIMDirectoryBearerTokenResponse, error) {
	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	scimDirID, err := idformat.SCIMDirectory.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	if _, err := q.ManagementGetSCIMDirectory(ctx, queries.ManagementGetSCIMDirectoryParams{
		EnvironmentID: envID,
		ID:            scimDirID,
	}); err != nil {
		return nil, fmt.Errorf("get scim directory: %w", err)
	}

	bearerToken := uuid.New()
	bearerTokenSHA := sha256.Sum256(bearerToken[:])

	if _, err := q.UpdateSCIMDirectoryBearerToken(ctx, queries.UpdateSCIMDirectoryBearerTokenParams{
		BearerTokenSha256: bearerTokenSHA[:],
		ID:                scimDirID,
	}); err != nil {
		return nil, fmt.Errorf("update scim directory access token: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ssoreadyv1.RotateSCIMDirectoryBearerTokenResponse{
		BearerToken: idformat.SCIMBearerToken.Format(bearerToken),
	}, nil
}
