package authninterceptor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/ssoready/ssoready/internal/authn"
	"github.com/ssoready/ssoready/internal/store"
)

var skipRPCs = []string{
	"/ssoready.v1.SSOReadyService/VerifyEmail",
	"/ssoready.v1.SSOReadyService/SignIn",
	"/ssoready.v1.SSOReadyService/AdminRedeemOneTimeToken",
}

var nonManagementAPIRPCs = []string{
	"/ssoready.v1.SSOReadyService/GetSAMLRedirectURL",
	"/ssoready.v1.SSOReadyService/RedeemSAMLAccessCode",
	"/ssoready.v1.SSOReadyService/ListSCIMUsers",
	"/ssoready.v1.SSOReadyService/GetSCIMUser",
	"/ssoready.v1.SSOReadyService/ListSCIMGroups",
	"/ssoready.v1.SSOReadyService/GetSCIMGroup",
}

var adminRPCs = []string{
	"/ssoready.v1.SSOReadyService/AdminWhoami",
	"/ssoready.v1.SSOReadyService/AdminCreateTestModeSAMLFlow",
	"/ssoready.v1.SSOReadyService/AdminListSAMLConnections",
	"/ssoready.v1.SSOReadyService/AdminGetSAMLConnection",
	"/ssoready.v1.SSOReadyService/AdminCreateSAMLConnection",
	"/ssoready.v1.SSOReadyService/AdminUpdateSAMLConnection",
	"/ssoready.v1.SSOReadyService/AdminParseSAMLMetadata",
	"/ssoready.v1.SSOReadyService/AdminListSAMLFlows",
	"/ssoready.v1.SSOReadyService/AdminGetSAMLFlow",
	"/ssoready.v1.SSOReadyService/AdminListSCIMDirectories",
	"/ssoready.v1.SSOReadyService/AdminGetSCIMDirectory",
	"/ssoready.v1.SSOReadyService/AdminCreateSCIMDirectory",
	"/ssoready.v1.SSOReadyService/AdminUpdateSCIMDirectory",
	"/ssoready.v1.SSOReadyService/AdminRotateSCIMDirectoryBearerToken",
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

			for _, rpc := range adminRPCs {
				if req.Spec().Procedure == rpc {
					res, err := s.AdminGetAdminSession(ctx, secretValue)
					if err != nil {
						return nil, fmt.Errorf("store: get admin session: %w", err)
					}

					ctx = authn.NewContext(ctx, authn.ContextData{
						AdminAccessToken: &authn.AdminAccessTokenData{
							OrganizationID: res.OrganizationID,
							CanManageSAML:  res.CanManageSAML,
							CanManageSCIM:  res.CanManageSCIM,
						},
					})
					return next(ctx, req)
				}
			}

			if strings.HasPrefix(secretValue, "ssoready_sk_") {
				// it's an api key
				apiKey, err := s.GetAPIKeyBySecretToken(ctx, &store.GetAPIKeyBySecretTokenRequest{Token: secretValue})
				if err != nil {
					return nil, err
				}

				// if it's not a management api key, make sure it's hitting an allowed endpoint
				if !apiKey.HasManagementAPIAccess {
					var ok bool
					for _, rpc := range nonManagementAPIRPCs {
						if req.Spec().Procedure == rpc {
							ok = true
						}
					}

					if !ok {
						return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("api key is not authorized to access management api"))
					}
				}

				ctx = authn.NewContext(ctx, authn.ContextData{
					APIKey: &authn.APIKeyData{
						AppOrgID: apiKey.AppOrganizationID,
						EnvID:    apiKey.EnvironmentID,
						APIKeyID: apiKey.ID,
					},
				})
			} else {
				session, err := s.GetAppSession(ctx, &store.GetAppSessionRequest{SessionToken: secretValue, Now: time.Now()})
				if err != nil {
					return nil, fmt.Errorf("appauthinterceptor: store: get app session: %w", err)
				}

				ctx = authn.NewContext(ctx, authn.ContextData{
					AppSession: &authn.AppSessionData{
						AppOrgID:     session.AppOrganizationID,
						AppUserID:    session.AppUserID,
						AppSessionID: session.AppSessionID,
					},
				})
			}

			return next(ctx, req)
		}
	}
}
