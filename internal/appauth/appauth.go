package appauth

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct{}

type ctxValue struct {
	OrgID     uuid.UUID
	APIKeyID  string
	AppUserID string
}

func WithAPIKey(ctx context.Context, orgID uuid.UUID, apiKeyID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, ctxValue{
		OrgID:    orgID,
		APIKeyID: apiKeyID,
	})
}

func WithAppUserID(ctx context.Context, orgID uuid.UUID, appUserID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, ctxValue{
		OrgID:     orgID,
		AppUserID: appUserID,
	})
}

func OrgID(ctx context.Context) uuid.UUID {
	return ctx.Value(ctxKey{}).(ctxValue).OrgID
}

func AppUserID(ctx context.Context) string {
	return ctx.Value(ctxKey{}).(ctxValue).AppUserID
}
