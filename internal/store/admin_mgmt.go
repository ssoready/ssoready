package store

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) CreateSetupURL(ctx context.Context, req *ssoreadyv1.CreateSetupURLRequest) (*ssoreadyv1.CreateSetupURLResponse, error) {
	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	orgID, err := idformat.Organization.Parse(req.OrganizationId)
	if err != nil {
		return nil, fmt.Errorf("parse organization id: %w", err)
	}

	// idor check
	org, err := q.ManagementGetOrganization(ctx, queries.ManagementGetOrganizationParams{
		EnvironmentID: envID,
		ID:            orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}

	oneTimeToken := uuid.New()
	oneTimeTokenSHA := sha256.Sum256(oneTimeToken[:])

	if _, err := q.CreateAdminAccessToken(ctx, queries.CreateAdminAccessTokenParams{
		ID:                 uuid.New(),
		OrganizationID:     org.ID,
		OneTimeTokenSha256: oneTimeTokenSHA[:],
		CreateTime:         time.Now(),
		ExpireTime:         time.Now().Add(time.Hour * 24),
		CanManageSaml:      &req.CanManageSaml,
		CanManageScim:      &req.CanManageScim,
	}); err != nil {
		return nil, fmt.Errorf("create admin access token: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	loginURL, err := url.Parse(s.defaultAdminSetupURL)
	if err != nil {
		panic(fmt.Errorf("parse default admin login url: %w", err))
	}

	query := url.Values{}
	query.Set("one-time-token", idformat.AdminOneTimeToken.Format(oneTimeToken))

	loginURL.RawQuery = query.Encode()

	return &ssoreadyv1.CreateSetupURLResponse{
		Url: loginURL.String(),
	}, nil
}
