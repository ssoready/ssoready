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
	"github.com/ssoready/ssoready/internal/authn"
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
	config := openidConfig{
		Issuer:                            fmt.Sprintf("%s/v1/oauth", s.BaseURL),
		AuthorizationEndpoint:             fmt.Sprintf("%s/v1/oauth/authorize", s.BaseURL),
		TokenEndpoint:                     fmt.Sprintf("%s/v1/oauth/token", s.BaseURL),
		JWKSURI:                           fmt.Sprintf("%s/v1/oauth/jwks", s.BaseURL),
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post"},
	}
	if err := json.NewEncoder(w).Encode(config); err != nil {
		panic(err)
	}
}

func (s *Service) oauthAuthorize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	clientID := r.URL.Query().Get("client_id")
	state := r.URL.Query().Get("state")
	orgID := r.URL.Query().Get("organization_id")
	orgExternalID := r.URL.Query().Get("organization_external_id")
	samlConnID := r.URL.Query().Get("saml_connection_id")

	slog.InfoContext(ctx, "oauth_authorize", "org_id", orgID, "org_external_id", orgExternalID, "saml_conn_id", samlConnID)

	getClientRes, err := s.Store.AuthOAuthGetClient(ctx, &store.AuthOAuthGetClientRequest{
		SAMLOAuthClientID: clientID,
	})
	if err != nil {
		panic(fmt.Errorf("get oauth client: %w", err))
	}

	ctx = authn.NewContext(ctx, authn.ContextData{
		SAMLOAuthClient: &authn.SAMLOAuthClientData{
			AppOrgID:      getClientRes.AppOrgID,
			EnvID:         getClientRes.EnvID,
			OAuthClientID: getClientRes.SAMLOAuthClientID,
		},
	})

	dataRes, err := s.Store.AuthGetOAuthAuthorizeData(ctx, &store.AuthGetOAuthAuthorizeDataRequest{
		OrganizationID:         orgID,
		OrganizationExternalID: orgExternalID,
		SAMLConnectionID:       samlConnID,
	})
	if err != nil {
		panic(fmt.Errorf("get oauth authorize data: %w", err))
	}

	slog.InfoContext(ctx, "oauth_authorize_saml_connection", "saml_connection_id", dataRes.SAMLConnectionID)

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
		panic(fmt.Errorf("upsert oauth authorize data: %w", err))
	}

	if err := acsTemplate.Execute(w, &acsTemplateData{
		SignOnURL:   dataRes.IDPRedirectURL,
		SAMLRequest: initRes.SAMLRequest,
		RelayState: s.StateSigner.Encode(statesign.Data{
			SAMLFlowID: samlFlowID,
		}),
	}); err != nil {
		panic(fmt.Errorf("acsTemplate.Execute: %w", err))
	}
}

type tokenResponse struct {
	IDToken string `json:"id_token"`
}

type idTokenClaims struct {
	jwt.Claims

	OrganizationID         string `json:"organizationId"`
	OrganizationExternalID string `json:"organizationExternalId"`
}

func (s *Service) oauthToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	getClientRes, err := s.Store.AuthOAuthGetClientWithSecret(ctx, &store.AuthOAuthGetClientWithSecretRequest{
		SAMLOAuthClientID:     clientID,
		SAMLOAuthClientSecret: clientSecret,
	})
	if err != nil {
		panic(fmt.Errorf("get oauth client: %w", err))
	}

	ctx = authn.NewContext(ctx, authn.ContextData{
		SAMLOAuthClient: &authn.SAMLOAuthClientData{
			AppOrgID:      getClientRes.AppOrgID,
			EnvID:         getClientRes.EnvID,
			OAuthClientID: getClientRes.SAMLOAuthClientID,
		},
	})

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
	claims := idTokenClaims{
		Claims: jwt.Claims{
			IssuedAt: jwt.NewNumericDate(now),
			Expiry:   jwt.NewNumericDate(now.Add(time.Hour)),
			Issuer:   fmt.Sprintf("%s/v1/oauth", s.BaseURL),
			Audience: jwt.Audience{getClientRes.SAMLOAuthClientID},

			Subject: res.Email,
		},
		OrganizationID:         res.OrganizationId,
		OrganizationExternalID: res.OrganizationExternalId,
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
