package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/appauth"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) RedeemSAMLAccessToken(ctx context.Context, req *ssoreadyv1.RedeemSAMLAccessTokenRequest) (*ssoreadyv1.RedeemSAMLAccessTokenResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	samlAccessToken, err := idformat.SAMLAccessToken.Parse(req.AccessToken)
	if err != nil {
		return nil, err
	}

	samlAccessTokenID := uuid.UUID(samlAccessToken) // todo ugh

	samlAccessTokenData, err := q.GetSAMLAccessTokenData(ctx, queries.GetSAMLAccessTokenDataParams{
		AppOrganizationID: appauth.OrgID(ctx),
		SecretAccessToken: &samlAccessTokenID,
	})
	if err != nil {
		return nil, fmt.Errorf("get saml access token data: %w", err)
	}

	var attrs map[string]string
	if err := json.Unmarshal(samlAccessTokenData.SubjectIdpAttributes, &attrs); err != nil {
		return nil, err
	}

	return &ssoreadyv1.RedeemSAMLAccessTokenResponse{
		SubjectIdpId:           *samlAccessTokenData.SubjectID,
		SubjectIdpAttributes:   attrs,
		OrganizationId:         idformat.Organization.Format(samlAccessTokenData.OrganizationID),
		OrganizationExternalId: derefOrEmpty(samlAccessTokenData.ExternalID),
		EnvironmentId:          idformat.Environment.Format(samlAccessTokenData.EnvironmentID),
	}, nil
}
