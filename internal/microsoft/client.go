package microsoft

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Client struct {
	HTTPClient                 *http.Client
	MicrosoftOAuthClientID     string
	MicrosoftOAuthClientSecret string
	MicrosoftOAuthRedirectURI  string
}

type Profile struct {
	TenantID string
	Name     string
	Email    string
}

type tokenResponse struct {
	IDToken string `json:"id_token"`
}

type tokenClaims struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	TID   string `json:"tid"`
	jwt.Claims
}

func (c *Client) ExchangeToken(ctx context.Context, code string) (*Profile, error) {
	reqBody := url.Values{}
	reqBody.Set("client_id", c.MicrosoftOAuthClientID)
	reqBody.Set("client_secret", c.MicrosoftOAuthClientSecret)
	reqBody.Set("redirect_uri", c.MicrosoftOAuthRedirectURI)
	reqBody.Set("scope", "openid profile email")
	reqBody.Set("grant_type", "authorization_code")
	reqBody.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://login.microsoftonline.com/common/oauth2/v2.0/token", strings.NewReader(reqBody.Encode()))
	if err != nil {
		return nil, fmt.Errorf("http: new request: %w", err)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: do request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("http: bad response status code: %v, reqBody: %s", res.StatusCode, string(body))
	}

	var tokenRes tokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tokenRes); err != nil {
		return nil, fmt.Errorf("parse response body: %w", err)
	}

	var parser jwt.Parser
	var claims tokenClaims
	if _, _, err := parser.ParseUnverified(tokenRes.IDToken, &claims); err != nil {
		return nil, fmt.Errorf("parse jwt claims: %w", err)
	}

	// For work and school accounts, the GUID is the immutable tenant ID of the organization that the user is signing in
	// to. For sign-ins to the personal Microsoft account tenant (services like Xbox, Teams for Life, or Outlook), the
	// value is 9188040d-6c67-4c5b-b112-36a304b66dad.
	//
	// https://learn.microsoft.com/en-us/entra/identity-platform/id-token-claims-reference
	var tenantID string
	if claims.TID != "9188040d-6c67-4c5b-b112-36a304b66dad" {
		tenantID = claims.TID
	}

	// sanity checks
	if claims.Name == "" || claims.Email == "" {
		return nil, fmt.Errorf("invalid claims, missing name or email: %v", claims)
	}

	return &Profile{
		TenantID: tenantID,
		Name:     claims.Name,
		Email:    claims.Email,
	}, nil
}
