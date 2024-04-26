package appauthinterceptor

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/ssoready/ssoready/internal/appauth"
	"github.com/ssoready/ssoready/internal/store"
)

func New(s *store.Store) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			authorization := req.Header().Get("Authorization")
			if authorization == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			secretValue, ok := strings.CutPrefix(authorization, "Bearer ")
			if !ok {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			apiKey, err := s.GetAPIKeyBySecretToken(ctx, &store.GetAPIKeyBySecretTokenRequest{Token: secretValue})
			if err != nil {
				return nil, err
			}

			ctx = appauth.WithAPIKey(ctx, apiKey.AppOrganizationID, apiKey.ID)
			return next(ctx, req)
		}
	}
}
