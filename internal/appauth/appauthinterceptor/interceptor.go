package appauthinterceptor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/ssoready/ssoready/internal/apikeyauth"
	"github.com/ssoready/ssoready/internal/appauth"
	"github.com/ssoready/ssoready/internal/store"
)

var skipRPCs = []string{
	"/ssoready.v1.SSOReadyService/VerifyEmail",
	"/ssoready.v1.SSOReadyService/SignIn",
}

func New(s *store.Store) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			for _, rpc := range skipRPCs {
				if req.Spec().Procedure == rpc {
					return next(ctx, req)
				}
			}

			authorization := req.Header().Get("Authorization")
			if authorization == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authorization header is required"))
			}

			secretValue, ok := strings.CutPrefix(authorization, "Bearer ")
			if !ok {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			if strings.HasPrefix(secretValue, "ssoready_sk_") {
				// it's an api key
				apiKey, err := s.GetAPIKeyBySecretToken(ctx, &store.GetAPIKeyBySecretTokenRequest{Token: secretValue})
				if err != nil {
					return nil, err
				}

				ctx = apikeyauth.WithAPIKey(ctx, apiKey.AppOrganizationID, apiKey.EnvironmentID)
			} else {
				session, err := s.GetAppSession(ctx, &store.GetAppSessionRequest{SessionToken: secretValue, Now: time.Now()})
				if err != nil {
					return nil, fmt.Errorf("appauthinterceptor: store: get app session: %w", err)
				}

				ctx = appauth.WithAppUserID(ctx, session.AppOrganizationID, session.AppUserID)
			}

			return next(ctx, req)
		}
	}
}
