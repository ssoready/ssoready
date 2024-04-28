package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
)

type Client struct {
	HTTPClient          *http.Client
	GoogleOAuthClientID string
}

type ParseCredentialRequest struct {
	Credential string
}

type ParseCredentialResponse struct {
	HostedDomain string
	Email        string
	Name         string
}

type googleCredentialClaims struct {
	jwt.Claims
	HostedDomain string `json:"hd"`
	Email        string `json:"email"`
	Name         string `json:"name"`
}

func (c *Client) ParseCredential(ctx context.Context, req *ParseCredentialRequest) (*ParseCredentialResponse, error) {
	token, err := jwt.ParseSigned(req.Credential)
	if err != nil {
		return nil, fmt.Errorf("parse credential: %w", err)
	}

	jwks, err := c.getGoogleOAuthJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("get google oauth jwks: %w", err)
	}

	var claims googleCredentialClaims
	if err := token.Claims(jwks, &claims); err != nil {
		return nil, fmt.Errorf("parse credential claims: %w", err)
	}

	if err := claims.Validate(jwt.Expected{
		Issuer:   "https://accounts.google.com",
		Audience: []string{c.GoogleOAuthClientID},
		Time:     time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("validate credential claims: %w", err)
	}

	return &ParseCredentialResponse{
		HostedDomain: claims.HostedDomain,
		Email:        claims.Email,
		Name:         claims.Name,
	}, nil
}

const googleOAuthJWKSURL = "https://www.googleapis.com/oauth2/v3/certs"

func (c *Client) getGoogleOAuthJWKS(ctx context.Context) (*jose.JSONWebKeySet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleOAuthJWKSURL, nil)
	if err != nil {
		return nil, fmt.Errorf("http: new request: %w", err)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: do request: %w", err)
	}

	defer res.Body.Close()

	var jwks jose.JSONWebKeySet
	if err := json.NewDecoder(res.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}

	return &jwks, nil
}
