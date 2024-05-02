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
		RequestID:      stateData.SAMLLoginEventID,
		IDPRedirectURL: *res.IdpRedirectUrl,
		SPEntityID:     *res.SpEntityID,
	}, nil
}

type AuthCreateInitiateTimelineEntryRequest struct {
	State       string
	InitiateURL string
}

func (s *Store) AuthCreateInitiateTimelineEntry(ctx context.Context, req *AuthCreateInitiateTimelineEntryRequest) error {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return err
	}
	defer rollback()

	stateData, err := s.statesigner.Decode(req.State)
	if err != nil {
		return err
	}

	samlLoginEventID, err := idformat.SAMLLoginEvent.Parse(stateData.SAMLLoginEventID)
	if err != nil {
		return err
	}

	if _, err := q.CreateSAMLLoginEventTimelineEntry(ctx, queries.CreateSAMLLoginEventTimelineEntryParams{
		ID:               uuid.New(),
		SamlLoginEventID: samlLoginEventID,
		Timestamp:        time.Now(),
		Type:             queries.SamlLoginEventTimelineEntryTypeSamlInitiate,
		SamlInitiateUrl:  &req.InitiateURL,
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
	SAMLLoginEventID     string
	SAMLConnectionID     string
	SubjectID            string
	SubjectIDPAttributes map[string]string
	RawSAMLPayload       string
}

type AuthUpsertSAMLLoginEventResponse struct {
	Token string
}

func (s *Store) AuthUpsertSAMLLoginEvent(ctx context.Context, req *AuthUpsertSAMLLoginEventRequest) (*AuthUpsertSAMLLoginEventResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	samlConnID, err := idformat.SAMLConnection.Parse(req.SAMLConnectionID)
	if err != nil {
		return nil, err
	}

	var samlLoginEventID uuid.UUID

	if req.SAMLLoginEventID == "" {
		if _, err := q.CreateSAMLLoginEvent(ctx, queries.CreateSAMLLoginEventParams{
			ID:               uuid.New(),
			SamlConnectionID: samlConnID,
			AccessCode:       uuid.New(),
			ExpireTime:       time.Now().Add(time.Hour),
		}); err != nil {
			return nil, err
		}
	} else {
		samlLoginEventID, err = idformat.SAMLLoginEvent.Parse(req.SAMLLoginEventID)
		if err != nil {
			return nil, err
		}
	}

	attrs, err := json.Marshal(req.SubjectIDPAttributes)
	if err != nil {
		return nil, err
	}

	qSAMLLoginEvent, err := q.UpdateSAMLLoginEventSubjectData(ctx, queries.UpdateSAMLLoginEventSubjectDataParams{
		ID:                   samlLoginEventID,
		SubjectIdpID:         &req.SubjectID,
		SubjectIdpAttributes: attrs,
	})
	if err != nil {
		return nil, err
	}

	// todo think through the security consequences here more deeply
	if qSAMLLoginEvent.SamlConnectionID != samlConnID {
		panic(fmt.Errorf("invariant failure: login_event.conn != conn: %q, %q", qSAMLLoginEvent.SamlConnectionID, req.SAMLConnectionID))
	}

	if _, err := q.CreateSAMLLoginEventTimelineEntry(ctx, queries.CreateSAMLLoginEventTimelineEntryParams{
		ID:                          uuid.New(),
		SamlLoginEventID:            qSAMLLoginEvent.ID,
		Timestamp:                   time.Now(),
		Type:                        queries.SamlLoginEventTimelineEntryTypeSamlReceiveAssertion,
		SamlReceiveAssertionPayload: &req.RawSAMLPayload,
	}); err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &AuthUpsertSAMLLoginEventResponse{
		Token: idformat.SAMLAccessCode.Format(qSAMLLoginEvent.AccessCode),
	}, nil
}
