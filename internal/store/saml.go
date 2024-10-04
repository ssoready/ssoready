package store

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ssoready/ssoready/internal/authn"
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

	authnData := authn.FullContextData(ctx)
	if authnData.APIKey == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("api key authentication is required"))
	}

	envID, err := idformat.Environment.Parse(authnData.APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	var samlConnID uuid.UUID
	if req.SamlConnectionId != "" {
		samlConnID, err = idformat.SAMLConnection.Parse(req.SamlConnectionId)
		if err != nil {
			return nil, err
		}
	} else if req.OrganizationId != "" {
		orgID, err := idformat.Organization.Parse(req.OrganizationId)
		if err != nil {
			return nil, err
		}

		samlConnID, err = q.GetPrimarySAMLConnectionIDByOrganizationID(ctx, queries.GetPrimarySAMLConnectionIDByOrganizationIDParams{
			EnvironmentID: envID,
			ID:            orgID,
		})
		if err != nil {
			return nil, err
		}
	} else if req.OrganizationExternalId != "" {
		samlConnID, err = q.GetPrimarySAMLConnectionIDByOrganizationExternalID(ctx, queries.GetPrimarySAMLConnectionIDByOrganizationExternalIDParams{
			EnvironmentID: envID,
			ExternalID:    &req.OrganizationExternalId,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("bad organization_external_id: organization not found, or organization does not have a primary SAML connection"))
			}
			return nil, err
		}
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("one of saml_connection_id, organization_id, or organization_external_id must be provided"))
	}

	samlRedirectData, err := q.GetSAMLRedirectURLData(ctx, queries.GetSAMLRedirectURLDataParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		EnvironmentID:     envID,
		SamlConnectionID:  samlConnID,
	})
	if err != nil {
		return nil, err
	}

	if samlRedirectData.IdpEntityID == nil || samlRedirectData.IdpRedirectUrl == nil || samlRedirectData.IdpX509Certificate == nil {
		// upsert an errored saml flow
		if _, err := s.upsertNotConfiguredSAMLFlow(ctx, q, &UpsertNotConfiguredSAMLFlowRequest{
			SAMLConnectionID:            req.SamlConnectionId,
			SAMLConnectionNotConfigured: true,
		}); err != nil {
			return nil, err
		}

		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("saml connection is not fully configured, see: https://ssoready.com/docs/ssoready-concepts/saml-connections#not-yet-configured"))
	}

	authURL := s.defaultAuthURL
	if samlRedirectData.EnvironmentAuthUrl != nil {
		authURL = *samlRedirectData.EnvironmentAuthUrl
	}

	samlFlowID := uuid.New()

	redirectURLQuery := url.Values{}
	redirectURLQuery.Set("state", s.statesigner.Encode(statesign.Data{
		SAMLFlowID: idformat.SAMLFlow.Format(samlFlowID),
	}))

	redirectURL, err := url.Parse(authURL)
	if err != nil {
		return nil, err
	}
	redirectURL = redirectURL.JoinPath(fmt.Sprintf("/v1/saml/%s/init", idformat.SAMLConnection.Format(samlConnID)))
	redirectURL.RawQuery = redirectURLQuery.Encode()

	redirect := redirectURL.String()

	now := time.Now()
	if _, err := q.CreateSAMLFlowGetRedirect(ctx, queries.CreateSAMLFlowGetRedirectParams{
		ID:               samlFlowID,
		SamlConnectionID: samlConnID,
		ExpireTime:       time.Now().Add(time.Hour),
		State:            req.State,
		CreateTime:       time.Now(),
		UpdateTime:       time.Now(),
		AuthRedirectUrl:  &redirect,
		GetRedirectTime:  &now,
		Status:           queries.SamlFlowStatusInProgress,
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

	var environmentID string
	authnData := authn.FullContextData(ctx)
	if authnData.APIKey != nil {
		environmentID = authnData.APIKey.EnvID
	} else if authnData.SAMLOAuthClient != nil {
		environmentID = authnData.SAMLOAuthClient.EnvID
	} else {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("api key or saml oauth client authentication is required"))
	}

	envID, err := idformat.Environment.Parse(environmentID)
	if err != nil {
		return nil, err
	}

	samlAccessCode, err := idformat.SAMLAccessCode.Parse(req.SamlAccessCode)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("bad saml_access_code: %w", err))
	}

	samlAccessCodeSHA := sha256.Sum256(samlAccessCode[:])

	samlAccessTokenData, err := q.GetSAMLAccessCodeData(ctx, queries.GetSAMLAccessCodeDataParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		EnvironmentID:     envID,
		AccessCodeSha256:  samlAccessCodeSHA[:],
	})
	if err != nil {
		return nil, fmt.Errorf("get saml access code data: %w", err)
	}

	var attrs map[string]string
	if err := json.Unmarshal(samlAccessTokenData.SubjectIdpAttributes, &attrs); err != nil {
		return nil, err
	}

	res := &ssoreadyv1.RedeemSAMLAccessCodeResponse{
		Email:                  *samlAccessTokenData.Email,
		Attributes:             attrs,
		State:                  samlAccessTokenData.State,
		OrganizationId:         idformat.Organization.Format(samlAccessTokenData.OrganizationID),
		OrganizationExternalId: derefOrEmpty(samlAccessTokenData.OrganizationExternalID),
		SamlFlowId:             idformat.SAMLFlow.Format(samlAccessTokenData.SamlFlowID),
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
		Status:         queries.SamlFlowStatusSucceeded,
	}); err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return res, nil
}

type UpsertNotConfiguredSAMLFlowRequest struct {
	SAMLConnectionID                         string
	SAMLConnectionNotConfigured              bool
	EnvironmentOAuthRedirectURINotConfigured bool
}

func (s *Store) UpsertNotConfiguredSAMLFlow(ctx context.Context, req *UpsertNotConfiguredSAMLFlowRequest) (string, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return "", err
	}
	defer rollback()

	id, err := s.upsertNotConfiguredSAMLFlow(ctx, q, req)
	if err != nil {
		return "", err
	}

	if err := commit(); err != nil {
		return "", err
	}

	return id, nil
}

func (s *Store) upsertNotConfiguredSAMLFlow(ctx context.Context, q *queries.Queries, req *UpsertNotConfiguredSAMLFlowRequest) (string, error) {
	samlConnID, err := idformat.SAMLConnection.Parse(req.SAMLConnectionID)
	if err != nil {
		return "", err
	}

	now := time.Now()
	qSAMLFlow, err := q.CreateSAMLFlowGetRedirect(ctx, queries.CreateSAMLFlowGetRedirectParams{
		ID:                               uuid.New(),
		SamlConnectionID:                 samlConnID,
		ExpireTime:                       time.Now().Add(time.Hour),
		State:                            "",
		CreateTime:                       time.Now(),
		UpdateTime:                       time.Now(),
		GetRedirectTime:                  &now,
		Status:                           queries.SamlFlowStatusFailed,
		ErrorSamlConnectionNotConfigured: req.SAMLConnectionNotConfigured,
		ErrorEnvironmentOauthRedirectUriNotConfigured: req.EnvironmentOAuthRedirectURINotConfigured,
	})
	if err != nil {
		return "", err
	}

	return idformat.SAMLFlow.Format(qSAMLFlow.ID), nil
}
