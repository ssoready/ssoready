package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) CreateAdminSetupURL(ctx context.Context, req *ssoreadyv1.CreateAdminSetupURLRequest) (*ssoreadyv1.CreateAdminSetupURLResponse, error) {
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
	org, err := q.GetOrganization(ctx, queries.GetOrganizationParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                orgID,
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

	return &ssoreadyv1.CreateAdminSetupURLResponse{
		Url: loginURL.String(),
	}, nil
}

func (s *Store) AdminRedeemOneTimeToken(ctx context.Context, req *ssoreadyv1.AdminRedeemOneTimeTokenRequest) (*ssoreadyv1.AdminRedeemOneTimeTokenResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	oneTimeToken, err := idformat.AdminOneTimeToken.Parse(req.OneTimeToken)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	oneTimeTokenSHA := sha256.Sum256(oneTimeToken[:])
	adminAccessToken, err := q.AdminGetAdminAccessTokenByOneTimeToken(ctx, oneTimeTokenSHA[:])
	if err != nil {
		return nil, fmt.Errorf("get admin access token: %w", err)
	}

	// generate token as 32-byte random string
	var tokenBytes [32]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return nil, err
	}
	accessTokenHex := hex.EncodeToString(tokenBytes[:])
	accessTokenSHA := sha256.Sum256(tokenBytes[:])

	if _, err := q.AdminConvertAdminAccessTokenToSession(ctx, queries.AdminConvertAdminAccessTokenToSessionParams{
		AccessTokenSha256: accessTokenSHA[:],
		ID:                adminAccessToken.ID,
	}); err != nil {
		return nil, fmt.Errorf("convert admin access token to session: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ssoreadyv1.AdminRedeemOneTimeTokenResponse{
		AdminSessionToken: accessTokenHex,
	}, nil
}
