package store

import (
	"context"
	"fmt"

	"github.com/ssoready/ssoready/internal/store/queries"
)

type UpdateAppOrganizationEntitlementsRequest struct {
	StripeCustomerID      string
	EntitledManagementAPI bool
	EntitledCustomDomains bool
}

func (s *Store) UpdateAppOrganizationEntitlements(ctx context.Context, req *UpdateAppOrganizationEntitlementsRequest) error {
	if _, err := s.q.UpdateAppOrganizationEntitlementsByStripeCustomerID(ctx, queries.UpdateAppOrganizationEntitlementsByStripeCustomerIDParams{
		StripeCustomerID:      &req.StripeCustomerID,
		EntitledManagementApi: &req.EntitledManagementAPI,
		EntitledCustomDomains: &req.EntitledCustomDomains,
	}); err != nil {
		return fmt.Errorf("update app organization entitlements by stripe customer id: %w", err)
	}
	return nil
}
