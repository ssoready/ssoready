package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

// todo break this code out from api's store layer, because the auth model is completely different

type AuthGetInitDataRequest struct {
	State            string
	SAMLConnectionID string
}

type AuthGetInitDataResponse struct {
	RequestID      string
	IDPRedirectURL string
	SPEntityID     string
}

func (s *Store) AuthGetInitData(ctx context.Context, req *AuthGetInitDataRequest) (*AuthGetInitDataResponse, error) {
	samlConnID, err := idformat.SAMLConnection.Parse(req.SAMLConnectionID)
	if err != nil {
		return nil, err
	}

	res, err := s.q.AuthGetInitData(ctx, samlConnID)
	if err != nil {
		return nil, err
	}

	stateData, err := s.statesigner.Decode(req.State)
	if err != nil {
		return nil, err
	}

	return &AuthGetInitDataResponse{
		RequestID:      stateData.SAMLFlowID,
		IDPRedirectURL: *res.IdpRedirectUrl,
		SPEntityID:     res.SpEntityID,
	}, nil
}

type AuthUpsertInitiateDataRequest struct {
	State           string
	InitiateRequest string
}

func (s *Store) AuthUpsertInitiateData(ctx context.Context, req *AuthUpsertInitiateDataRequest) error {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return err
	}
	defer rollback()

	stateData, err := s.statesigner.Decode(req.State)
	if err != nil {
		return err
	}

	samlFlowID, err := idformat.SAMLFlow.Parse(stateData.SAMLFlowID)
	if err != nil {
		return err
	}

	qSAMLFlow, err := q.AuthGetSAMLFlow(ctx, samlFlowID)
	if err != nil {
		return err
	}

	now := time.Now()
	if _, err := q.UpsertSAMLFlowInitiate(ctx, queries.UpsertSAMLFlowInitiateParams{
		ID:               samlFlowID,
		SamlConnectionID: qSAMLFlow.SamlConnectionID,
		ExpireTime:       time.Now().Add(time.Hour),
		State:            stateData.State,
		CreateTime:       time.Now(),
		UpdateTime:       time.Now(),
		InitiateRequest:  &req.InitiateRequest,
		InitiateTime:     &now,
		Status:           queries.SamlFlowStatusInProgress,
	}); err != nil {
		return err
	}

	if err := commit(); err != nil {
		return err
	}

	return nil
}

type AuthGetValidateDataRequest struct {
	SAMLConnectionID string
}

type AuthGetValidateDataResponse struct {
	SPEntityID             string
	IDPEntityID            string
	IDPX509Certificate     []byte
	OrganizationDomains    []string
	EnvironmentRedirectURL string
}

func (s *Store) AuthGetValidateData(ctx context.Context, req *AuthGetValidateDataRequest) (*AuthGetValidateDataResponse, error) {
	samlConnID, err := idformat.SAMLConnection.Parse(req.SAMLConnectionID)
	if err != nil {
		return nil, err
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	res, err := q.AuthGetValidateData(ctx, samlConnID)
	if err != nil {
		return nil, err
	}

	domains, err := q.AuthGetSAMLConnectionDomains(ctx, samlConnID)
	if err != nil {
		return nil, err
	}

	return &AuthGetValidateDataResponse{
		SPEntityID:             res.SpEntityID,
		IDPEntityID:            *res.IdpEntityID,
		IDPX509Certificate:     res.IdpX509Certificate,
		OrganizationDomains:    domains,
		EnvironmentRedirectURL: *res.RedirectUrl,
	}, nil
}

func (s *Store) AuthCheckAssertionAlreadyProcessed(ctx context.Context, samlFlowID string) (bool, error) {
	if samlFlowID == "" {
		return false, nil
	}

	id, err := idformat.SAMLFlow.Parse(samlFlowID)
	if err != nil {
		return false, err
	}

	ok, err := s.q.AuthCheckAssertionAlreadyProcessed(ctx, id)
	if err != nil {
		return false, err
	}

	return ok, nil
}

type AuthUpsertSAMLLoginEventRequest struct {
	SAMLConnectionID                     string
	SAMLFlowID                           string
	Email                                string
	SubjectIDPAttributes                 map[string]string
	SAMLAssertion                        string
	ErrorUnsignedAssertion               bool
	ErrorBadIssuer                       *string
	ErrorBadAudience                     *string
	ErrorBadSubjectID                    *string
	ErrorEmailOutsideOrganizationDomains *string
}

type AuthUpsertSAMLLoginEventResponse struct {
	SAMLFlowID string
	Token      string
}

func (s *Store) AuthUpsertReceiveAssertionData(ctx context.Context, req *AuthUpsertSAMLLoginEventRequest) (*AuthUpsertSAMLLoginEventResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	samlConnID, err := idformat.SAMLConnection.Parse(req.SAMLConnectionID)
	if err != nil {
		return nil, err
	}

	var samlFlowID uuid.UUID
	if req.SAMLFlowID != "" {
		samlFlowID, err = idformat.SAMLFlow.Parse(req.SAMLFlowID)
		if err != nil {
			return nil, err
		}
	} else {
		samlFlowID = uuid.New()
	}

	var assertionOk bool
	if !req.ErrorUnsignedAssertion && req.ErrorBadIssuer == nil && req.ErrorBadAudience == nil && req.ErrorBadSubjectID == nil && req.ErrorEmailOutsideOrganizationDomains == nil {
		assertionOk = true
	}

	// create a new flow
	now := time.Now()

	var accessCode *uuid.UUID
	status := queries.SamlFlowStatusFailed
	if assertionOk {
		id := uuid.New()
		accessCode = &id
		status = queries.SamlFlowStatusInProgress
	}

	qSAMLFlow, err := q.UpsertSAMLFlowReceiveAssertion(ctx, queries.UpsertSAMLFlowReceiveAssertionParams{
		ID:                                   samlFlowID,
		SamlConnectionID:                     samlConnID,
		AccessCode:                           accessCode,
		ExpireTime:                           time.Now().Add(time.Hour),
		State:                                "",
		CreateTime:                           time.Now(),
		UpdateTime:                           time.Now(),
		Assertion:                            &req.SAMLAssertion,
		ReceiveAssertionTime:                 &now,
		ErrorUnsignedAssertion:               req.ErrorUnsignedAssertion,
		ErrorBadIssuer:                       req.ErrorBadIssuer,
		ErrorBadAudience:                     req.ErrorBadAudience,
		ErrorBadSubjectID:                    req.ErrorBadSubjectID,
		ErrorEmailOutsideOrganizationDomains: req.ErrorEmailOutsideOrganizationDomains,
		Status:                               status,
	})
	if err != nil {
		return nil, err
	}

	if qSAMLFlow.SamlConnectionID != samlConnID {
		return nil, fmt.Errorf("saml flow does not belong to given saml connection")
	}

	attrs, err := json.Marshal(req.SubjectIDPAttributes)
	if err != nil {
		return nil, err
	}

	if _, err := q.UpdateSAMLFlowSubjectData(ctx, queries.UpdateSAMLFlowSubjectDataParams{
		ID:                   samlFlowID,
		Email:                &req.Email,
		SubjectIdpAttributes: attrs,
	}); err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	var token string
	if qSAMLFlow.AccessCode != nil {
		token = idformat.SAMLAccessCode.Format(*qSAMLFlow.AccessCode)
	}

	return &AuthUpsertSAMLLoginEventResponse{
		SAMLFlowID: idformat.SAMLFlow.Format(qSAMLFlow.ID),
		Token:      token,
	}, nil
}

type AuthUpdateAppRedirectURLRequest struct {
	SAMLFlowID     string
	AppRedirectURL string
}

func (s *Store) AuthUpdateAppRedirectURL(ctx context.Context, req *AuthUpdateAppRedirectURLRequest) error {
	samlFlowID, err := idformat.SAMLFlow.Parse(req.SAMLFlowID)
	if err != nil {
		return err
	}

	if _, err := s.q.UpdateSAMLFlowAppRedirectURL(ctx, queries.UpdateSAMLFlowAppRedirectURLParams{
		ID:             samlFlowID,
		AppRedirectUrl: &req.AppRedirectURL,
	}); err != nil {
		return err
	}

	return nil
}
