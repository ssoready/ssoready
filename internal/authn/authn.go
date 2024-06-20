package authn

import (
	"context"

	"github.com/google/uuid"
)

type ContextData struct {
	AppSession      *AppSessionData
	APIKey          *APIKeyData
	SAMLOAuthClient *SAMLOAuthClientData
}

type AppSessionData struct {
	AppOrgID     uuid.UUID
	AppUserID    string
	AppSessionID string
}

type APIKeyData struct {
	AppOrgID uuid.UUID
	EnvID    string
	APIKeyID string
}

type SAMLOAuthClientData struct {
	AppOrgID      uuid.UUID
	EnvID         string
	OAuthClientID string
}

type ctxKey struct{}

func AppOrgID(ctx context.Context) uuid.UUID {
	v, ok := ctx.Value(ctxKey{}).(ContextData)
	if !ok {
		panic("ctx does not carry ContextData")
	}

	if v.AppSession != nil {
		return v.AppSession.AppOrgID
	}
	if v.APIKey != nil {
		return v.APIKey.AppOrgID
	}
	if v.SAMLOAuthClient != nil {
		return v.SAMLOAuthClient.AppOrgID
	}
	panic("invalid ContextData")
}

func FullContextData(ctx context.Context) ContextData {
	v, ok := ctx.Value(ctxKey{}).(ContextData)
	if !ok {
		panic("ctx does not carry ContextData")
	}
	return v
}

func NewContext(ctx context.Context, data ContextData) context.Context {
	return context.WithValue(ctx, ctxKey{}, data)
}
