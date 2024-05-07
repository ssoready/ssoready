package store

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/appauth"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/statesign"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) GetSAMLRedirectURL(ctx context.Context, req *ssoreadyv1.GetSAMLRedirectURLRequest) (*ssoreadyv1.GetSAMLRedirectURLResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
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

	samlFlowID := uuid.New()

	redirectURLQuery := url.Values{}
	redirectURLQuery.Set("state", s.statesigner.Encode(statesign.Data{
		SAMLFlowID: idformat.SAMLFlow.Format(samlFlowID),
		State:      req.State,
	}))

	redirectURL, err := url.Parse(authURL)
	if err != nil {
		return nil, err
	}
	redirectURL = redirectURL.JoinPath(fmt.Sprintf("/saml/%s/init", idformat.SAMLConnection.Format(samlConnID)))
	redirectURL.RawQuery = redirectURLQuery.Encode()

	redirect := redirectURL.String()

	now := time.Now()
	if _, err := q.CreateSAMLFlowGetRedirect(ctx, queries.CreateSAMLFlowGetRedirectParams{
		ID:               samlFlowID,
		SamlConnectionID: samlConnID,
		AccessCode:       uuid.New(),
		ExpireTime:       time.Now().Add(time.Hour),
		State:            req.State,
		CreateTime:       time.Now(),
		UpdateTime:       time.Now(),
		AuthRedirectUrl:  &redirect,
		GetRedirectTime:  &now,
	}); err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &ssoreadyv1.GetSAMLRedirectURLResponse{RedirectUrl: redirect}, nil
}

func (s *Store) RedeemSAMLAccessCode(ctx context.Context, req *ssoreadyv1.RedeemSAMLAccessCodeRequest) (*ssoreadyv1.RedeemSAMLAccessCodeResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	samlAccessCode, err := idformat.SAMLAccessCode.Parse(req.AccessCode)
	if err != nil {
		return nil, err
	}

	samlAccessTokenData, err := q.GetSAMLAccessCodeData(ctx, queries.GetSAMLAccessCodeDataParams{
		AppOrganizationID: appauth.OrgID(ctx),
		AccessCode:        samlAccessCode,
	})
	if err != nil {
		return nil, fmt.Errorf("get saml access code data: %w", err)
	}

	var attrs map[string]string
	if err := json.Unmarshal(samlAccessTokenData.SubjectIdpAttributes, &attrs); err != nil {
		return nil, err
	}

	res := &ssoreadyv1.RedeemSAMLAccessCodeResponse{
		SubjectIdpId:           *samlAccessTokenData.SubjectIdpID,
		SubjectIdpAttributes:   attrs,
		State:                  samlAccessTokenData.State,
		OrganizationId:         idformat.Organization.Format(samlAccessTokenData.OrganizationID),
		OrganizationExternalId: derefOrEmpty(samlAccessTokenData.OrganizationExternalID),
		EnvironmentId:          idformat.Environment.Format(samlAccessTokenData.EnvironmentID),
	}

	resJSON, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}

	now := time.Now()
	if _, err := q.UpdateSAMLFlowRedeem(ctx, queries.UpdateSAMLFlowRedeemParams{
		ID:             samlAccessTokenData.SamlFlowID,
		UpdateTime:     time.Now(),
		RedeemTime:     &now,
		RedeemResponse: resJSON,
	}); err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return res, nil
}
