package store

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ssoready/ssoready/internal/authn"
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

type AuthGetInitDataBadStateError struct {
	err error
}

func (e *AuthGetInitDataBadStateError) Error() string {
	return fmt.Sprintf("bad state param: %s", e.err.Error())
}

func (s *Store) AuthGetInitData(ctx context.Context, req *AuthGetInitDataRequest) (*AuthGetInitDataResponse, error) {
	samlConnID, err := idformat.SAMLConnection.Parse(req.SAMLConnectionID)
	if err != nil {
		return nil, fmt.Errorf("parse saml connection id: %w", err)
	}

	res, err := s.q.AuthGetInitData(ctx, samlConnID)
	if err != nil {
		return nil, fmt.Errorf("get init data: %w", err)
	}

	stateData, err := s.statesigner.Decode(req.State)
	if err != nil {
		return nil, &AuthGetInitDataBadStateError{err}
	}

	samlFlowID, err := idformat.SAMLFlow.Parse(stateData.SAMLFlowID)
	if err != nil {
		return nil, fmt.Errorf("parse saml flow id: %w", err)
	}

	qSAMLFlow, err := s.q.GetSAMLFlowByID(ctx, samlFlowID)
	if err != nil {
		return nil, fmt.Errorf("get saml flow by id: %w", err)
	}

	if qSAMLFlow.SamlConnectionID != samlConnID {
		return nil, fmt.Errorf("saml flow id does not match saml connection id")
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
	SPEntityID                  string
	IDPEntityID                 string
	IDPX509Certificate          []byte
	OrganizationDomains         []string
	EnvironmentRedirectURL      string
	EnvironmentOAuthRedirectURI string
	EnvironmentAdminTestModeURL string
}

var ErrNoSuchSAMLConnection = errors.New("no such saml connection")

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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoSuchSAMLConnection
		}

		return nil, fmt.Errorf("get validate data: %w", err)
	}

	domains, err := q.AuthGetSAMLConnectionDomains(ctx, samlConnID)
	if err != nil {
		return nil, fmt.Errorf("get domains: %w", err)
	}

	if res.IdpEntityID == nil || res.IdpX509Certificate == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("saml connection is not fully configured, see: https://ssoready.com/docs/ssoready-concepts/saml-connections#not-yet-configured"))
	}

	var adminTestModeURL string
	if res.AdminUrl != nil {
		adminTestModeURL = fmt.Sprintf("%s/test-mode", *res.AdminUrl)
	} else {
		adminTestModeURL = s.defaultAdminTestModeURL
	}

	return &AuthGetValidateDataResponse{
		SPEntityID:                  res.SpEntityID,
		IDPEntityID:                 *res.IdpEntityID,
		IDPX509Certificate:          res.IdpX509Certificate,
		OrganizationDomains:         domains,
		EnvironmentRedirectURL:      *res.RedirectUrl,
		EnvironmentOAuthRedirectURI: derefOrEmpty(res.OauthRedirectUri),
		EnvironmentAdminTestModeURL: adminTestModeURL,
	}, nil
}

var InvalidSAMLRequestID = errors.New("invalid saml request id")

func (s *Store) AuthCheckAssertionAlreadyProcessed(ctx context.Context, samlFlowID string) (bool, error) {
	if samlFlowID == "" {
		return false, nil
	}

	id, err := idformat.SAMLFlow.Parse(samlFlowID)
	if err != nil {
		return false, InvalidSAMLRequestID
	}

	ok, err := s.q.AuthCheckAssertionAlreadyProcessed(ctx, id)
	if err != nil {
		return false, err
	}

	return ok, nil
}

type AuthUpsertSAMLLoginEventRequest struct {
	SAMLConnectionID                     string
	SAMLAssertionID                      *string
	SAMLFlowID                           string
	Email                                string
	SubjectIDPAttributes                 map[string]string
	SAMLAssertion                        string
	ErrorSAMLConnectionNotConfigured     bool
	ErrorUnsignedAssertion               bool
	ErrorBadIssuer                       *string
	ErrorBadAudience                     *string
	ErrorBadSignatureAlgorithm           *string
	ErrorBadDigestAlgorithm              *string
	ErrorBadCertificate                  *x509.Certificate
	ErrorBadSubjectID                    *string
	ErrorEmailOutsideOrganizationDomains *string
}

type AuthUpsertSAMLLoginEventResponse struct {
	SAMLFlowID          string
	SAMLFlowIsOAuth     bool
	SAMLFlowTestModeIDP string
	Token               string
	State               string // only useful for oauth flow, where state must be returned at same time as code
}

var ErrDuplicateAssertionID = errors.New("an assertion with this ID has already been processed")
var ErrSAMLConnectionIDMismatch = errors.New("saml connection id does not match saml flow id")

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

	if req.SAMLFlowID != "" {
		samlFlowID, err := idformat.SAMLFlow.Parse(req.SAMLFlowID)
		if err != nil {
			return nil, err
		}

		qSAMLFlow, err := q.GetSAMLFlowByID(ctx, samlFlowID)
		if err != nil {
			return nil, err
		}

		if qSAMLFlow.SamlConnectionID != samlConnID {
			return nil, fmt.Errorf("saml flow id does not match saml connection id")
		}
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
	if !req.ErrorSAMLConnectionNotConfigured && !req.ErrorUnsignedAssertion && req.ErrorBadIssuer == nil && req.ErrorBadAudience == nil && req.ErrorBadSignatureAlgorithm == nil && req.ErrorBadDigestAlgorithm == nil && req.ErrorBadCertificate == nil && req.ErrorBadSubjectID == nil && req.ErrorEmailOutsideOrganizationDomains == nil {
		assertionOk = true
	}

	// create a new flow
	now := time.Now()

	var accessCode *uuid.UUID
	var accessCodeSHA []byte
	status := queries.SamlFlowStatusFailed
	if assertionOk {
		id := uuid.New()
		sha := sha256.Sum256(id[:])

		accessCode = &id
		accessCodeSHA = sha[:]
		status = queries.SamlFlowStatusInProgress
	}

	var badX509Certificate []byte
	if req.ErrorBadCertificate != nil {
		badX509Certificate = req.ErrorBadCertificate.Raw
	}

	qSAMLFlow, err := q.UpsertSAMLFlowReceiveAssertion(ctx, queries.UpsertSAMLFlowReceiveAssertionParams{
		ID:                                   samlFlowID,
		AssertionID:                          req.SAMLAssertionID,
		SamlConnectionID:                     samlConnID,
		AccessCodeSha256:                     accessCodeSHA,
		ExpireTime:                           time.Now().Add(time.Hour),
		State:                                "",
		CreateTime:                           time.Now(),
		UpdateTime:                           time.Now(),
		Assertion:                            &req.SAMLAssertion,
		ReceiveAssertionTime:                 &now,
		ErrorSamlConnectionNotConfigured:     req.ErrorSAMLConnectionNotConfigured,
		ErrorUnsignedAssertion:               req.ErrorUnsignedAssertion,
		ErrorBadIssuer:                       req.ErrorBadIssuer,
		ErrorBadAudience:                     req.ErrorBadAudience,
		ErrorBadSignatureAlgorithm:           req.ErrorBadSignatureAlgorithm,
		ErrorBadDigestAlgorithm:              req.ErrorBadDigestAlgorithm,
		ErrorBadX509Certificate:              badX509Certificate,
		ErrorBadSubjectID:                    req.ErrorBadSubjectID,
		ErrorEmailOutsideOrganizationDomains: req.ErrorEmailOutsideOrganizationDomains,
		Status:                               status,
	})
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			if pgxErr.Code == "23505" && pgxErr.ConstraintName == "saml_flows_saml_connection_id_assertion_id_key" {
				return nil, ErrDuplicateAssertionID
			}
		}

		return nil, err
	}

	if qSAMLFlow.SamlConnectionID != samlConnID {
		return nil, ErrSAMLConnectionIDMismatch
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
	if accessCode != nil {
		token = idformat.SAMLAccessCode.Format(*accessCode)
	}

	return &AuthUpsertSAMLLoginEventResponse{
		SAMLFlowID:          idformat.SAMLFlow.Format(qSAMLFlow.ID),
		SAMLFlowIsOAuth:     qSAMLFlow.IsOauth != nil && *qSAMLFlow.IsOauth,
		SAMLFlowTestModeIDP: derefOrEmpty(qSAMLFlow.TestModeIdp),
		Token:               token,
		State:               qSAMLFlow.State,
	}, nil
}

type AuthGetOAuthAuthorizeDataRequest struct {
	OrganizationID         string
	OrganizationExternalID string
	SAMLConnectionID       string
}

type AuthGetOAuthAuthorizeDataResponse struct {
	IDPRedirectURL   string
	SPEntityID       string
	SAMLConnectionID string
}

func (s *Store) AuthGetOAuthAuthorizeData(ctx context.Context, req *AuthGetOAuthAuthorizeDataRequest) (*AuthGetOAuthAuthorizeDataResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	authnData := authn.FullContextData(ctx)
	if authnData.SAMLOAuthClient == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("oauth authentication is required"))
	}

	envID, err := idformat.Environment.Parse(authnData.SAMLOAuthClient.EnvID)
	if err != nil {
		return nil, err
	}

	var samlConnID uuid.UUID
	if req.SAMLConnectionID != "" {
		samlConnID, err = idformat.SAMLConnection.Parse(req.SAMLConnectionID)
		if err != nil {
			return nil, err
		}
	} else if req.OrganizationID != "" {
		orgID, err := idformat.Organization.Parse(req.OrganizationID)
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
	} else if req.OrganizationExternalID != "" {
		samlConnID, err = q.GetPrimarySAMLConnectionIDByOrganizationExternalID(ctx, queries.GetPrimarySAMLConnectionIDByOrganizationExternalIDParams{
			EnvironmentID: envID,
			ExternalID:    &req.OrganizationExternalID,
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

	samlConn, err := q.GetSAMLConnectionByID(ctx, samlConnID)
	if err != nil {
		return nil, err
	}

	qEnv, err := q.GetEnvironmentByID(ctx, envID)
	if err != nil {
		return nil, err
	}

	if derefOrEmpty(qEnv.OauthRedirectUri) == "" {
		// even when we fail in this way, give the resolved saml connection id
		// back for logging a failed saml flow
		return &AuthGetOAuthAuthorizeDataResponse{
			SAMLConnectionID: idformat.SAMLConnection.Format(samlConn.ID),
		}, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("environment OAuth redirect URI not configured, see: https://ssoready.com/docs/ssoready-concepts/saml-login-flows#environment-oauth-redirect-uri-not-configured"))
	}

	if samlConn.IdpEntityID == nil || samlConn.IdpRedirectUrl == nil || samlConn.IdpX509Certificate == nil {
		// even when we fail in this way, give the resolved saml connection id
		// back for logging a failed saml flow
		return &AuthGetOAuthAuthorizeDataResponse{
			SAMLConnectionID: idformat.SAMLConnection.Format(samlConn.ID),
		}, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("saml connection is not fully configured, see: https://ssoready.com/docs/ssoready-concepts/saml-flows#saml-connection-not-fully-configured"))
	}

	return &AuthGetOAuthAuthorizeDataResponse{
		IDPRedirectURL:   *samlConn.IdpRedirectUrl,
		SPEntityID:       samlConn.SpEntityID,
		SAMLConnectionID: idformat.SAMLConnection.Format(samlConn.ID),
	}, nil
}

type AuthUpsertOAuthAuthorizeDataRequest struct {
	SAMLFlowID       string
	SAMLConnectionID string
	State            string
	InitiateRequest  string
}

func (s *Store) AuthUpsertOAuthAuthorizeData(ctx context.Context, req *AuthUpsertOAuthAuthorizeDataRequest) error {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return err
	}
	defer rollback()

	samlFlowID, err := idformat.SAMLFlow.Parse(req.SAMLFlowID)
	if err != nil {
		return err
	}

	samlConnectionID, err := idformat.SAMLConnection.Parse(req.SAMLConnectionID)
	if err != nil {
		return err
	}

	now := time.Now()
	isOAuth := true
	if _, err := q.UpsertSAMLFlowInitiate(ctx, queries.UpsertSAMLFlowInitiateParams{
		ID:               samlFlowID,
		SamlConnectionID: samlConnectionID,
		ExpireTime:       time.Now().Add(time.Hour),
		State:            req.State,
		CreateTime:       time.Now(),
		UpdateTime:       time.Now(),
		InitiateRequest:  &req.InitiateRequest,
		InitiateTime:     &now,
		Status:           queries.SamlFlowStatusInProgress,
		IsOauth:          &isOAuth,
	}); err != nil {
		return err
	}

	if err := commit(); err != nil {
		return err
	}

	return nil
}
