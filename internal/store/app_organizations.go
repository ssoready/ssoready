package store

import (
	"context"
	"fmt"

	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
)

func (s *Store) GetAppOrganization(ctx context.Context, req *ssoreadyv1.GetAppOrganizationRequest) (*ssoreadyv1.GetAppOrganizationResponse, error) {
	qAppOrg, err := s.q.GetAppOrganizationByID(ctx, authn.AppOrgID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get app org by id: %w", err)
	}

	return &ssoreadyv1.GetAppOrganizationResponse{
		GoogleHostedDomain: derefOrEmpty(qAppOrg.GoogleHostedDomain),
	}, nil
}
