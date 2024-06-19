package authservice

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/ssoready/ssoready/internal/apikeyauth"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/saml"
	"github.com/ssoready/ssoready/internal/statesign"
	"github.com/ssoready/ssoready/internal/store"
	"github.com/ssoready/ssoready/internal/store/idformat"
)

type openidConfig struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	JWKSURI                           string   `json:"jwks_uri"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
}

func (s *Service) oauthOpenIDConfiguration(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["org_id"]

	config := openidConfig{
		Issuer:                            fmt.Sprintf("http://localhost:8080/v1/oauth/%s", orgID),
		AuthorizationEndpoint:             fmt.Sprintf("http://localhost:8080/v1/oauth/%s/authorize", orgID),
		TokenEndpoint:                     fmt.Sprintf("http://localhost:8080/v1/oauth/%s/token", orgID),
		JWKSURI:                           fmt.Sprintf("http://localhost:8080/v1/oauth/%s/jwks", orgID),
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post"},
	}
	if err := json.NewEncoder(w).Encode(config); err != nil {
		panic(err)
	}
}

func (s *Service) oauthAuthorize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mux.Vars(r)["org_id"]
	state := r.URL.Query().Get("state")

	slog.InfoContext(ctx, "oauth_authorize", "org_id", orgID)

	dataRes, err := s.Store.AuthGetOAuthAuthorizeData(ctx, &store.AuthGetOAuthAuthorizeDataRequest{
		OrganizationID: orgID,
	})
	if err != nil {
		panic(err)
	}

	// In the standard flow, samlFlowID is assigned by GetRedirectURL. In the OAuth flow it's assigned here because
	// there is no "get redirect url" equivalent.
	samlFlowID := idformat.SAMLFlow.Format(uuid.New())

	initRes := saml.Init(&saml.InitRequest{
		RequestID:  samlFlowID,
		SPEntityID: dataRes.SPEntityID,
		Now:        time.Now(),
	})

	if err := s.Store.AuthUpsertOAuthAuthorizeData(ctx, &store.AuthUpsertOAuthAuthorizeDataRequest{
		State:            state,
		InitiateRequest:  initRes.InitiateRequest,
		SAMLConnectionID: dataRes.SAMLConnectionID,
		SAMLFlowID:       samlFlowID,
	}); err != nil {
		panic(err)
	}

	if err := acsTemplate.Execute(w, &acsTemplateData{
		SignOnURL:   dataRes.IDPRedirectURL,
		SAMLRequest: initRes.SAMLRequest,
		RelayState: s.StateSigner.Encode(statesign.Data{
			SAMLFlowID: samlFlowID,
			State:      r.FormValue("state"),
		}),
	}); err != nil {
		panic(fmt.Errorf("acsTemplate.Execute: %w", err))
	}
}

type tokenResponse struct {
	//AccessToken string `json:"access_token"`
	IDToken string `json:"id_token"`
}

func (s *Service) oauthToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	apiKey, err := s.Store.GetAPIKeyBySecretToken(ctx, &store.GetAPIKeyBySecretTokenRequest{Token: r.FormValue("client_secret")})
	if err != nil {
		panic(err)
	}

	ctx = apikeyauth.WithAPIKey(ctx, apiKey.AppOrganizationID, apiKey.EnvironmentID)
	res, err := s.Store.RedeemSAMLAccessCode(ctx, &ssoreadyv1.RedeemSAMLAccessCodeRequest{
		SamlAccessCode: r.FormValue("code"),
	})
	if err != nil {
		panic(err)
	}

	signerOptions := jose.SignerOptions{}
	signer, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       s.OAuthIDTokenPrivateKey,
	}, signerOptions.WithType("JWT"))
	if err != nil {
		panic(err)
	}

	now := time.Now()
	claims := jwt.Claims{
		IssuedAt: jwt.NewNumericDate(now),
		Expiry:   jwt.NewNumericDate(now.Add(time.Hour)),
		Issuer:   fmt.Sprintf("http://localhost:8080/v1/oauth/%s", res.OrganizationId),
		Audience: jwt.Audience{r.FormValue("client_id")},

		Subject: res.Email,
	}

	idToken, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
	if err != nil {
		panic(err)
	}

	if err := json.NewEncoder(w).Encode(tokenResponse{IDToken: idToken}); err != nil {
		panic(err)
	}
}

func (s *Service) oauthJWKS(w http.ResponseWriter, r *http.Request) {
	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Key: &s.OAuthIDTokenPrivateKey.PublicKey,
			},
		},
	}

	if err := json.NewEncoder(w).Encode(jwks); err != nil {
		panic(err)
	}
}
