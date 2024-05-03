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
		SPEntityID:     *res.SpEntityID,
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
		AccessCode:       uuid.New(),
		ExpireTime:       time.Now().Add(time.Hour),
		State:            stateData.State,
		CreateTime:       time.Now(),
		UpdateTime:       time.Now(),
		InitiateRequest:  &req.InitiateRequest,
		InitiateTime:     &now,
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
	EnvironmentRedirectURL string
}

func (s *Store) AuthGetValidateData(ctx context.Context, req *AuthGetValidateDataRequest) (*AuthGetValidateDataResponse, error) {
	samlConnID, err := idformat.SAMLConnection.Parse(req.SAMLConnectionID)
	if err != nil {
		return nil, err
	}

	res, err := s.q.AuthGetValidateData(ctx, samlConnID)
	if err != nil {
		return nil, err
	}

	return &AuthGetValidateDataResponse{
		SPEntityID:             *res.SpEntityID,
		IDPEntityID:            *res.IdpEntityID,
		IDPX509Certificate:     res.IdpX509Certificate,
		EnvironmentRedirectURL: *res.RedirectUrl,
	}, nil
}

type AuthUpsertSAMLLoginEventRequest struct {
	SAMLConnectionID     string
	SAMLFlowID           string
	SubjectID            string
	SubjectIDPAttributes map[string]string
	SAMLAssertion        string
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

	// create a new flow
	now := time.Now()
	qSAMLFlow, err := q.UpsertSAMLFlowReceiveAssertion(ctx, queries.UpsertSAMLFlowReceiveAssertionParams{
		ID:                   samlFlowID,
		SamlConnectionID:     samlConnID,
		AccessCode:           uuid.New(),
		ExpireTime:           time.Now().Add(time.Hour),
		CreateTime:           time.Now(),
		UpdateTime:           time.Now(),
		Assertion:            &req.SAMLAssertion,
		ReceiveAssertionTime: &now,
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
		SubjectIdpID:         &req.SubjectID,
		SubjectIdpAttributes: attrs,
	}); err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &AuthUpsertSAMLLoginEventResponse{
		SAMLFlowID: idformat.SAMLFlow.Format(qSAMLFlow.ID),
		Token:      idformat.SAMLAccessCode.Format(qSAMLFlow.AccessCode),
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
