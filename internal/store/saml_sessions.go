package store

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/appauth"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
	"golang.org/x/crypto/nacl/auth"
)

func (s *Store) GetSAMLRedirectURL(ctx context.Context, req *ssoreadyv1.GetSAMLRedirectURLRequest) (*ssoreadyv1.GetSAMLRedirectURLResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	samlConnID, err := idformat.SAMLConnection.Parse(req.SamlConnectionId)
	if err != nil {
		return nil, err
	}

	envAuthURL, err := q.GetSAMLRedirectURLData(ctx, queries.GetSAMLRedirectURLDataParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                samlConnID,
	})
	if err != nil {
		return nil, err
	}

	authURL := s.globalDefaultAuthURL
	if envAuthURL != nil {
		authURL = *envAuthURL
	}

	sig := auth.Sum([]byte(req.State), &s.samlStateSigningKey)
	state := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString([]byte(req.State)), base64.RawURLEncoding.EncodeToString(sig[:]))

	redirectURL, err := url.Parse(authURL)
	if err != nil {
		return nil, err
	}
	redirectURL = redirectURL.JoinPath(fmt.Sprintf("/saml/%s/init", idformat.SAMLConnection.Format(samlConnID)))
	redirectURL.RawQuery = url.Values{"state": []string{state}}.Encode()

	return &ssoreadyv1.GetSAMLRedirectURLResponse{RedirectUrl: redirectURL.String()}, nil
}

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
