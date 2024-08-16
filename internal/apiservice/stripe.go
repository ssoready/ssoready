package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/stripe/stripe-go/v79"
)

func (s *Service) GetStripeCheckoutURL(ctx context.Context, req *connect.Request[ssoreadyv1.GetStripeCheckoutURLRequest]) (*connect.Response[ssoreadyv1.GetStripeCheckoutURLResponse], error) {
	stripeCustomerID, err := s.Store.GetAppOrganizationStripeCustomerID(ctx)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	var customer *string
	if stripeCustomerID != "" {
		customer = &stripeCustomerID
	}

	checkoutSession, err := s.StripeClient.CheckoutSessions.New(&stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}", s.StripeCheckoutSuccessURL)),
		Customer:   customer,
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(s.StripePriceIDProTier),
				Quantity: stripe.Int64(1),
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return connect.NewResponse(&ssoreadyv1.GetStripeCheckoutURLResponse{
		Url: checkoutSession.URL,
	}), nil
}

func (s *Service) RedeemStripeCheckout(ctx context.Context, req *connect.Request[ssoreadyv1.RedeemStripeCheckoutRequest]) (*connect.Response[ssoreadyv1.RedeemStripeCheckoutResponse], error) {
	checkoutSession, err := s.StripeClient.CheckoutSessions.Get(req.Msg.StripeCheckoutSessionId, nil)
	if err != nil {
		panic(err)
	}

	if checkoutSession.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		panic("session not paid")
	}

	if err := s.Store.UpdateAppOrganizationStripeCustomerID(ctx, checkoutSession.Customer.ID); err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(&ssoreadyv1.RedeemStripeCheckoutResponse{}), nil
}

func (s *Service) GetStripeBillingPortalURL(ctx context.Context, req *connect.Request[ssoreadyv1.GetStripeBillingPortalURLRequest]) (*connect.Response[ssoreadyv1.GetStripeBillingPortalURLResponse], error) {
	stripeCustomerID, err := s.Store.GetAppOrganizationStripeCustomerID(ctx)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	if stripeCustomerID == "" {
		panic(fmt.Errorf("no stripe customer id"))
	}

	billingPortalSession, err := s.StripeClient.BillingPortalSessions.New(&stripe.BillingPortalSessionParams{
		Customer:  &stripeCustomerID,
		ReturnURL: &s.StripeBillingPortalReturnURL,
	})
	if err != nil {
		panic(err)
	}

	return connect.NewResponse(&ssoreadyv1.GetStripeBillingPortalURLResponse{
		Url: billingPortalSession.URL,
	}), nil
}
