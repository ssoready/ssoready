package appauth

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct{}

type ctxValue struct {
	OrgID     uuid.UUID
	AppUserID string
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

func MaybeOrgID(ctx context.Context) *uuid.UUID {
	v, ok := ctx.Value(ctxKey{}).(ctxValue)
	if !ok {
		return nil
	}
	return &v.OrgID
}

func MaybeAppUserID(ctx context.Context) *string {
	v, ok := ctx.Value(ctxKey{}).(ctxValue)
	if !ok {
		return nil
	}
	return &v.AppUserID
}
