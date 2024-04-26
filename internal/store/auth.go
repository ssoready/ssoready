package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/store/queries"
)

type GetSAMLConnectionByIDRequest struct {
	ID string
}

type GetSAMLConnectionByIDResponse struct {
	SAMLConnection *queries.SamlConnection
}

func (s *Store) GetSAMLConnectionByID(ctx context.Context, req *GetSAMLConnectionByIDRequest) (*GetSAMLConnectionByIDResponse, error) {
	samlConn, err := s.q.GetSAMLConnectionByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &GetSAMLConnectionByIDResponse{SAMLConnection: &samlConn}, nil
}

type GetOrganizationByIDRequest struct {
	ID string
}

type GetOrganizationByIDResponse struct {
	Organization *queries.Organization
}

func (s *Store) GetOrganizationByID(ctx context.Context, req *GetOrganizationByIDRequest) (*GetOrganizationByIDResponse, error) {
	org, err := s.q.GetOrganizationByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &GetOrganizationByIDResponse{Organization: &org}, nil
}

type GetEnvironmentByIDRequest struct {
	ID string
}

type GetEnvironmentByIDResponse struct {
	Environment *queries.Environment
}

func (s *Store) GetEnvironmentByID(ctx context.Context, req *GetEnvironmentByIDRequest) (*GetEnvironmentByIDResponse, error) {
	env, err := s.q.GetEnvironmentByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &GetEnvironmentByIDResponse{Environment: &env}, nil
}

type CreateSAMLSessionRequest struct {
	SAMLSession queries.SamlSession
}

type CreateSAMLSessionResponse struct {
	SAMLSession queries.SamlSession
}

func (s *Store) CreateSAMLSession(ctx context.Context, req *CreateSAMLSessionRequest) (*CreateSAMLSessionResponse, error) {
	secretAccessToken := uuid.NewString()
	samlSess, err := s.q.CreateSAMLSession(ctx, queries.CreateSAMLSessionParams{
		ID:                   uuid.NewString(),
		SamlConnectionID:     req.SAMLSession.SamlConnectionID,
		SecretAccessToken:    &secretAccessToken,
		SubjectID:            req.SAMLSession.SubjectID,
		SubjectIdpAttributes: req.SAMLSession.SubjectIdpAttributes,
	})
	if err != nil {
		return nil, err
	}

	return &CreateSAMLSessionResponse{SAMLSession: samlSess}, nil
}
