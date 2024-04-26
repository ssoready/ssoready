package store

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

// todo break this code out from api's store layer, because the auth model is completely different

type GetSAMLConnectionByIDRequest struct {
	ID string
}

func (s *Store) GetSAMLConnectionByID(ctx context.Context, req *GetSAMLConnectionByIDRequest) (*ssoreadyv1.SAMLConnection, error) {
	id, err := idformat.SAMLConnection.Parse(req.ID)
	if err != nil {
		return nil, err
	}

	samlConn, err := s.q.GetSAMLConnectionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &ssoreadyv1.SAMLConnection{
		Id:                 idformat.SAMLConnection.Format(samlConn.ID),
		OrganizationId:     idformat.Organization.Format(samlConn.OrganizationID),
		IdpRedirectUrl:     *samlConn.IdpRedirectUrl,
		IdpX509Certificate: samlConn.IdpX509Certificate,
		IdpEntityId:        *samlConn.IdpEntityID,
	}, nil
}

type GetOrganizationByIDRequest struct {
	ID string
}

func (s *Store) GetOrganizationByID(ctx context.Context, req *GetOrganizationByIDRequest) (*ssoreadyv1.Organization, error) {
	id, err := idformat.Organization.Parse(req.ID)
	if err != nil {
		return nil, err
	}

	org, err := s.q.GetOrganizationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &ssoreadyv1.Organization{
		Id:            idformat.Organization.Format(org.ID),
		EnvironmentId: idformat.Environment.Format(org.EnvironmentID),
	}, nil
}

type GetEnvironmentByIDRequest struct {
	ID string
}

func (s *Store) GetEnvironmentByID(ctx context.Context, req *GetEnvironmentByIDRequest) (*ssoreadyv1.Environment, error) {
	id, err := idformat.Environment.Parse(req.ID)
	if err != nil {
		return nil, err
	}

	env, err := s.q.GetEnvironmentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &ssoreadyv1.Environment{
		Id:          idformat.Environment.Format(env.ID),
		RedirectUrl: *env.RedirectUrl,
	}, nil
}

type CreateSAMLSessionRequest struct {
	SAMLConnectionID     string
	SubjectID            string
	SubjectIDPAttributes map[string]string
}

type CreateSAMLSessionResponse struct {
	Token string
}

func (s *Store) CreateSAMLSession(ctx context.Context, req *CreateSAMLSessionRequest) (*CreateSAMLSessionResponse, error) {
	samlConnID, err := idformat.SAMLConnection.Parse(req.SAMLConnectionID)
	if err != nil {
		return nil, err
	}

	attrs, err := json.Marshal(req.SubjectIDPAttributes)
	if err != nil {
		return nil, err
	}

	secretAccessToken := uuid.New()
	samlSess, err := s.q.CreateSAMLSession(ctx, queries.CreateSAMLSessionParams{
		ID:                   uuid.New(),
		SamlConnectionID:     samlConnID,
		SecretAccessToken:    &secretAccessToken,
		SubjectID:            &req.SubjectID,
		SubjectIdpAttributes: attrs,
	})
	if err != nil {
		return nil, err
	}

	return &CreateSAMLSessionResponse{Token: idformat.SAMLAccessToken.Format(*samlSess.SecretAccessToken)}, nil
}
