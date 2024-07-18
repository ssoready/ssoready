package apiservice

import (
	"net/http"

	"github.com/resend/resend-go/v2"
	"github.com/ssoready/ssoready/internal/gen/ssoready/v1/ssoreadyv1connect"
	"github.com/ssoready/ssoready/internal/google"
	"github.com/ssoready/ssoready/internal/microsoft"
	"github.com/ssoready/ssoready/internal/store"
)

type Service struct {
	Store                     *store.Store
	GoogleClient              *google.Client
	MicrosoftClient           *microsoft.Client
	ResendClient              *resend.Client
	EmailChallengeFrom        string
	EmailVerificationEndpoint string
	SAMLMetadataHTTPClient    *http.Client
	ssoreadyv1connect.UnimplementedSSOReadyServiceHandler
}
