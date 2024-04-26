package appauth

import (
	"context"
)

type ctxKey struct{}

type ctxValue struct {
	OrgID    string
	APIKeyID string
}

func WithAPIKey(ctx context.Context, orgID, apiKeyID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, ctxValue{
		OrgID:    orgID,
		APIKeyID: apiKeyID,
	})
}

func OrgID(ctx context.Context) string {
	return ctx.Value(ctxKey{}).(ctxValue).OrgID
}
