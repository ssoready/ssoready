package apistripewebhook

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/ssoready/ssoready/internal/store"
	"github.com/stripe/stripe-go/v79"
	stripeclient "github.com/stripe/stripe-go/v79/client"
	"github.com/stripe/stripe-go/v79/webhook"
)

type Service struct {
	Store                *store.Store
	StripeClient         *stripeclient.API
	StripeEndpointSecret string
}

func (s *Service) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		defer r.Body.Close()
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), s.StripeEndpointSecret)
		if err != nil {
			panic(err)
		}

		slog.InfoContext(ctx, "stripe_webhook_event", "type", event.Type)

		if event.Type != "entitlements.active_entitlement_summary.updated" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var entitlementSummary stripe.EntitlementsActiveEntitlementSummary
		if err := json.Unmarshal(event.Data.Raw, &entitlementSummary); err != nil {
			panic(err)
		}

		listEntitlementsIter := s.StripeClient.EntitlementsActiveEntitlements.List(&stripe.EntitlementsActiveEntitlementListParams{
			Customer: &entitlementSummary.Customer,
		})

		var hasManagementAPI, hasCustomDomains bool
		for listEntitlementsIter.Next() {
			entitlement := listEntitlementsIter.EntitlementsActiveEntitlement()
			if entitlement.LookupKey == "ssoready-management-api" {
				hasManagementAPI = true
			}
			if entitlement.LookupKey == "ssoready-custom-domains" {
				hasCustomDomains = true
			}
		}

		if err := s.Store.UpdateAppOrganizationEntitlements(ctx, &store.UpdateAppOrganizationEntitlementsRequest{
			StripeCustomerID:      entitlementSummary.Customer,
			EntitledManagementAPI: hasManagementAPI,
			EntitledCustomDomains: hasCustomDomains,
		}); err != nil {
			panic(err)
		}

		slog.InfoContext(ctx, "update_app_org_entitlements", "stripe_customer_id", entitlementSummary.Customer, "entitled_management_api", hasManagementAPI)

		w.WriteHeader(http.StatusOK)
	}
}
