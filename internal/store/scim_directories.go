package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) CreateSCIMDirectory(ctx context.Context, req *ssoreadyv1.CreateSCIMDirectoryRequest) (*ssoreadyv1.SCIMDirectory, error) {
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
	org, err := q.GetOrganization(ctx, queries.GetOrganizationParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                orgID,
	})
	if err != nil {
		return nil, err
	}

	env, err := q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                org.EnvironmentID,
	})
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
		ID:             uuid.New(),
		OrganizationID: orgID,
		IsPrimary:      req.ScimDirectory.Primary,
		ScimBaseUrl:    scimBaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create scim directory: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return parseSCIMDirectory(qSCIMDirectory), nil
}

func parseSCIMDirectory(qSCIMDirectory queries.ScimDirectory) *ssoreadyv1.SCIMDirectory {
	return &ssoreadyv1.SCIMDirectory{
		Id:                idformat.SCIMDirectory.Format(qSCIMDirectory.ID),
		OrganizationId:    idformat.Organization.Format(qSCIMDirectory.OrganizationID),
		Primary:           qSCIMDirectory.IsPrimary,
		ScimBaseUrl:       qSCIMDirectory.ScimBaseUrl,
		ClientBearerToken: "", // intentionally left blank
	}
}