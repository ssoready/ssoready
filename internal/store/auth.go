package store

import (
	"context"

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
