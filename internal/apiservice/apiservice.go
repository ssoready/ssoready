package apiservice

import (
	"net/http"

	"github.com/resend/resend-go/v2"
	"github.com/ssoready/ssoready/internal/gen/ssoready/v1/ssoreadyv1connect"
	"github.com/ssoready/ssoready/internal/google"
	"github.com/ssoready/ssoready/internal/microsoft"
	"github.com/ssoready/ssoready/internal/store"
	stripe "github.com/stripe/stripe-go/v79/client"
)

type Service struct {
	Store                        *store.Store
	GoogleClient                 *google.Client
	MicrosoftClient              *microsoft.Client
	ResendClient                 *resend.Client
	EmailChallengeFrom           string
	EmailVerificationEndpoint    string
	SAMLMetadataHTTPClient       *http.Client
	StripeClient                 *stripe.API
	StripeCheckoutSuccessURL     string
	StripePriceIDProTier         string
	StripeBillingPortalReturnURL string
	ssoreadyv1connect.UnimplementedSSOReadyServiceHandler
}
