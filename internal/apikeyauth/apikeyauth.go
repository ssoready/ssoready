package apikeyauth

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct{}

type ctxValue struct {
	OrgID uuid.UUID
	EnvID uuid.UUID
}

func WithAPIKey(ctx context.Context, orgID, envID uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKey{}, ctxValue{
		OrgID: orgID,
		EnvID: envID,
	})
}

func AppOrgID(ctx context.Context) uuid.UUID {
	return ctx.Value(ctxKey{}).(ctxValue).OrgID
}

func EnvID(ctx context.Context) uuid.UUID {
	return ctx.Value(ctxKey{}).(ctxValue).EnvID
}
