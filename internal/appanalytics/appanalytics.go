package appanalytics

import (
	"context"

	"connectrpc.com/connect"
	"github.com/segmentio/analytics-go/v3"
	"github.com/ssoready/ssoready/internal/authn"
)

func NewInterceptor(client analytics.Client) connect.UnaryInterceptorFunc {
	return func(f connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			ctx = newContext(ctx, client)
			return f(ctx, req)
		}
	}
}

func Track(ctx context.Context, event string, properties analytics.Properties) error {
	authnData := authn.FullContextData(ctx)
	if authnData.AppSession == nil {
		return nil // not a user session, so do no tracking
	}

	properties["app_organization_id"] = authn.AppOrgID(ctx).String()
	return FromContext(ctx).Enqueue(analytics.Track{
		Event:      event,
		Properties: properties,
		UserId:     authnData.AppSession.AppUserID,
	})
}

type ctxKey struct{}

func newContext(ctx context.Context, client analytics.Client) context.Context {
	return context.WithValue(ctx, ctxKey{}, client)
}

func FromContext(ctx context.Context) analytics.Client {
	return ctx.Value(ctxKey{}).(analytics.Client)
}

type NoopClient struct{}

func (n NoopClient) Close() error {
	return nil
}

func (n NoopClient) Enqueue(_ analytics.Message) error {
	return nil
}
