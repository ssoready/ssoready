package store

import (
	"context"
	"fmt"

	"github.com/ssoready/ssoready/internal/authn"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) GetAppOrganizationStripeCustomerID(ctx context.Context) (string, error) {
	qAppOrg, err := s.q.GetAppOrganizationByID(ctx, authn.AppOrgID(ctx))
	if err != nil {
		return "", fmt.Errorf("get app organization by id: %w", err)
	}

	return derefOrEmpty(qAppOrg.StripeCustomerID), nil
}

func (s *Store) UpdateAppOrganizationStripeCustomerID(ctx context.Context, stripeCustomerID string) error {
	if err := s.q.UpdateAppOrganizationStripeCustomerID(ctx, queries.UpdateAppOrganizationStripeCustomerIDParams{
		StripeCustomerID: &stripeCustomerID,
		ID:               authn.AppOrgID(ctx),
	}); err != nil {
		return fmt.Errorf("update app organization stripe customer id: %w", err)
	}
	return nil
}
