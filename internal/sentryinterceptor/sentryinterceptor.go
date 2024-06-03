package sentryinterceptor

import (
	"context"
	"encoding/json"
	"errors"

	"connectrpc.com/connect"
	"github.com/getsentry/sentry-go"
	"github.com/ssoready/ssoready/internal/appauth"
	"google.golang.org/protobuf/encoding/protojson"
)

func NewPreAuthentication() connect.UnaryInterceptorFunc {
	return func(f connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			sentry.GetHubFromContext(ctx).ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTag("endpoint", req.Spec().Procedure)
			})

			res, err := f(ctx, req)
			if err != nil {
				// add connectrpc's error details onto Sentry error context
				var connectErr *connect.Error
				if errors.As(err, &connectErr) {
					// this process is a bit of a hack; convert details to json and back
					var details []any
					for _, d := range connectErr.Details() {
						v, err := d.Value()
						if err != nil {
							return nil, err
						}

						vjson, err := protojson.Marshal(v)
						if err != nil {
							return nil, err
						}

						var vdata any
						if err := json.Unmarshal(vjson, &vdata); err != nil {
							return nil, err
						}

						details = append(details, map[string]any{
							"type":  d.Type(),
							"value": vdata,
						})
					}

					sentry.GetHubFromContext(ctx).ConfigureScope(func(scope *sentry.Scope) {
						scope.SetContext("connect_err", map[string]any{
							"details": details,
						})
					})
				}

				sentry.GetHubFromContext(ctx).CaptureException(err)
			}

			return res, err
		}
	}
}

func NewPostAuthentication() connect.UnaryInterceptorFunc {
	return func(f connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			appOrgID := appauth.MaybeOrgID(ctx)
			appUserID := appauth.MaybeAppUserID(ctx)
			sentry.GetHubFromContext(ctx).ConfigureScope(func(scope *sentry.Scope) {
				if appOrgID != nil {
					scope.SetTag("org_id", appOrgID.String())
				}
				if appUserID != nil {
					scope.SetUser(sentry.User{
						ID: *appUserID,
					})
				}
			})

			return f(ctx, req)
		}
	}
}
