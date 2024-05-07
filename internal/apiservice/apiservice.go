package apiservice

import (
	"github.com/resend/resend-go/v2"
	"github.com/ssoready/ssoready/internal/gen/ssoready/v1/ssoreadyv1connect"
	"github.com/ssoready/ssoready/internal/google"
	"github.com/ssoready/ssoready/internal/store"
)

type Service struct {
	Store                     *store.Store
	GoogleClient              *google.Client
	ResendClient              *resend.Client
	EmailChallengeFrom        string
	EmailVerificationEndpoint string
	ssoreadyv1connect.UnimplementedSSOReadyServiceHandler
}
