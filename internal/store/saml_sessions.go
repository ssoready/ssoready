package store

import (
	"context"
	"encoding/json"

	"github.com/ssoready/ssoready/internal/appauth"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) RedeemSAMLAccessToken(ctx context.Context, req *ssoreadyv1.RedeemSAMLAccessTokenRequest) (*ssoreadyv1.RedeemSAMLAccessTokenResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	samlAccessTokenData, err := q.GetSAMLAccessTokenData(ctx, queries.GetSAMLAccessTokenDataParams{
		AppOrganizationID: appauth.OrgID(ctx),
		SecretAccessToken: &req.AccessToken,
	})
	if err != nil {
		return nil, err
	}

	var attrs map[string]string
	if err := json.Unmarshal(samlAccessTokenData.SubjectIdpAttributes, &attrs); err != nil {
		return nil, err
	}

	return &ssoreadyv1.RedeemSAMLAccessTokenResponse{
		SubjectIdpId:         *samlAccessTokenData.SubjectID,
		SubjectIdpAttributes: attrs,
		OrganizationId:       samlAccessTokenData.OrganizationID,
		EnvironmentId:        samlAccessTokenData.EnvironmentID,
	}, nil
}
