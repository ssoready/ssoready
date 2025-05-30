package authservice

import (
	"crypto/rsa"
	"crypto/x509"
	"embed"
	_ "embed"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/gorilla/mux"
	"github.com/ssoready/ssoready/internal/emailaddr"
	"github.com/ssoready/ssoready/internal/saml"
	"github.com/ssoready/ssoready/internal/statesign"
	"github.com/ssoready/ssoready/internal/store"
)

type acsTemplateData struct {
	SignOnURL   string
	SAMLRequest string
}

var acsTemplate = template.Must(template.New("acs").Parse(`
<html>
	<body>
		<form method="POST" action="{{ .SignOnURL }}">
			<input type="hidden" name="SAMLRequest" value="{{ .SAMLRequest }}"></input>
		</form>
		<script>
			document.forms[0].submit();
		</script>
	</body>
</html>
`))

type errorTemplateData struct {
	ErrorMessage            string
	SAMLFlowID              string
	WantIDPEntityID         string
	GotIDPEntityID          string
	WantAudienceRestriction string
	GotAudienceRestriction  string
	GotSignatureAlgorithm   string
	GotDigestAlgorithm      string
	WantCertificatePEM      string
	GotCertificatePEM       string
	GotSubjectID            string
	WantEmailDomains        string
}

//go:embed templates/static
var staticData embed.FS
var staticFS, _ = fs.Sub(staticData, "templates/static")

//go:embed templates/error.html
var errorTemplateContent string
var errorTemplate = template.Must(template.New("error").Parse(errorTemplateContent))

type Service struct {
	BaseURL                string
	Store                  *store.Store
	OAuthIDTokenPrivateKey *rsa.PrivateKey
	StateSigner            statesign.Signer
}

func (s *Service) NewHandler() http.Handler {
	r := mux.NewRouter()

	r.PathPrefix("/internal/static/").Handler(http.StripPrefix("/internal/static/", http.FileServer(http.FS(staticFS))))
	r.Handle("/v1/saml/{saml_conn_id}/init", logHandlerNoRespHeaders(http.HandlerFunc(s.samlInit))).Methods("GET")
	r.Handle("/v1/saml/{saml_conn_id}/acs", logHandlerNoRespHeaders(http.HandlerFunc(s.samlAcs))).Methods("POST")

	r.HandleFunc("/v1/oauth/.well-known/openid-configuration", s.oauthOpenIDConfiguration).Methods("GET")
	r.HandleFunc("/v1/oauth/authorize", s.oauthAuthorize).Methods("GET")
	r.HandleFunc("/v1/oauth/token", s.oauthToken).Methods("POST")
	r.HandleFunc("/v1/oauth/userinfo", s.oauthUserinfo).Methods("GET")
	r.HandleFunc("/v1/oauth/jwks", s.oauthJWKS).Methods("GET")

	r.Handle("/v1/scim/{scim_directory_id}/Users", logHandlerIncludeRespHeaders(s.scimMiddleware(s.scimListUsers))).Methods(http.MethodGet)
	r.Handle("/v1/scim/{scim_directory_id}/Users/{scim_user_id}", logHandlerIncludeRespHeaders(s.scimMiddleware(s.scimGetUser))).Methods(http.MethodGet)
	r.Handle("/v1/scim/{scim_directory_id}/Users", logHandlerIncludeRespHeaders(s.scimMiddleware(s.scimCreateUser))).Methods(http.MethodPost)
	r.Handle("/v1/scim/{scim_directory_id}/Users/{scim_user_id}", logHandlerIncludeRespHeaders(s.scimMiddleware(s.scimUpdateUser))).Methods(http.MethodPut)
	r.Handle("/v1/scim/{scim_directory_id}/Users/{scim_user_id}", logHandlerIncludeRespHeaders(s.scimMiddleware(s.scimPatchUser))).Methods(http.MethodPatch)
	r.Handle("/v1/scim/{scim_directory_id}/Users/{scim_user_id}", logHandlerIncludeRespHeaders(s.scimMiddleware(s.scimDeleteUser))).Methods(http.MethodDelete)
	r.Handle("/v1/scim/{scim_directory_id}/Groups/{scim_group_id}", logHandlerIncludeRespHeaders(s.scimMiddleware(s.scimGetGroup))).Methods(http.MethodGet)
	r.Handle("/v1/scim/{scim_directory_id}/Groups", logHandlerIncludeRespHeaders(s.scimMiddleware(s.scimListGroups))).Methods(http.MethodGet)
	r.Handle("/v1/scim/{scim_directory_id}/Groups", logHandlerIncludeRespHeaders(s.scimMiddleware(s.scimCreateGroup))).Methods(http.MethodPost)
	r.Handle("/v1/scim/{scim_directory_id}/Groups/{scim_group_id}", logHandlerIncludeRespHeaders(s.scimMiddleware(s.scimUpdateGroup))).Methods(http.MethodPut)
	r.Handle("/v1/scim/{scim_directory_id}/Groups/{scim_group_id}", logHandlerIncludeRespHeaders(s.scimMiddleware(s.scimPatchGroup))).Methods(http.MethodPatch)
	r.Handle("/v1/scim/{scim_directory_id}/Groups/{scim_group_id}", logHandlerIncludeRespHeaders(s.scimMiddleware(s.scimDeleteGroup))).Methods(http.MethodDelete)

	return r
}

func (s *Service) samlInit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	samlConnID := mux.Vars(r)["saml_conn_id"]
	state := r.URL.Query().Get("state")

	slog.InfoContext(ctx, "init", "saml_connection_id", samlConnID, "state", state)

	dataRes, err := s.Store.AuthGetInitData(ctx, &store.AuthGetInitDataRequest{
		SAMLConnectionID: samlConnID,
		State:            state,
	})
	if err != nil {
		var badState *store.AuthGetInitDataBadStateError
		if errors.As(err, &badState) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		panic(err)
	}

	initRes := saml.Init(&saml.InitRequest{
		RequestID:  dataRes.RequestID,
		SPEntityID: dataRes.SPEntityID,
		Now:        time.Now(),
	})

	if err := s.Store.AuthUpsertInitiateData(ctx, &store.AuthUpsertInitiateDataRequest{
		State:           state,
		InitiateRequest: initRes.InitiateRequest,
	}); err != nil {
		panic(err)
	}

	if err := acsTemplate.Execute(w, &acsTemplateData{
		SignOnURL:   dataRes.IDPRedirectURL,
		SAMLRequest: initRes.SAMLRequest,
	}); err != nil {
		panic(fmt.Errorf("acsTemplate.Execute: %w", err))
	}
}

func (s *Service) samlAcs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	samlConnID := mux.Vars(r)["saml_conn_id"]

	slog.InfoContext(ctx, "acs", "saml_connection_id", samlConnID)

	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	slog.InfoContext(ctx, "acs_form", "form", r.Form)

	dataRes, err := s.Store.AuthGetValidateData(ctx, &store.AuthGetValidateDataRequest{
		SAMLConnectionID: samlConnID,
	})
	if err != nil {
		if errors.Is(err, store.ErrNoSuchSAMLConnection) {
			http.Error(w, "saml connection not found", http.StatusNotFound)
			return
		}

		var connectErr *connect.Error
		if errors.As(err, &connectErr) && connectErr.Code() == connect.CodeFailedPrecondition {
			createSAMLLoginRes, err := s.Store.AuthUpsertReceiveAssertionData(ctx, &store.AuthUpsertSAMLLoginEventRequest{
				SAMLConnectionID:                 samlConnID,
				ErrorSAMLConnectionNotConfigured: true,
			})
			if err != nil {
				panic(err)
			}

			if err := errorTemplate.Execute(w, &errorTemplateData{
				ErrorMessage: "SAML connection is not fully configured. This needs to be fixed in the Service Provider.",
				SAMLFlowID:   createSAMLLoginRes.SAMLFlowID,
			}); err != nil {
				panic(fmt.Errorf("acsTemplate.Execute: %w", err))
			}
			return
		}

		panic(err)
	}

	cert, err := x509.ParseCertificate(dataRes.IDPX509Certificate)
	if err != nil {
		panic(err)
	}

	validateRes, err := saml.Validate(&saml.ValidateRequest{
		SAMLResponse:   r.FormValue("SAMLResponse"),
		IDPCertificate: cert,
		IDPEntityID:    dataRes.IDPEntityID,
		SPEntityID:     dataRes.SPEntityID,
		Now:            time.Now(),
	})

	slog.InfoContext(ctx, "acs_validate", "validate_res", validateRes, "validate_err", err)

	var validateFailed bool // used as a failsafe later

	// populated even in the case of validate errors
	var (
		requestID   string
		assertionID string
		assertion   string
	)

	// populated when there are validate errors
	var (
		malformedAssertion    bool
		unsignedAssertion     bool
		expiredAssertion      bool
		badIssuer             *string
		badAudience           *string
		badSignatureAlgorithm *string
		badDigestAlgorithm    *string
		badCertificate        *x509.Certificate
	)

	// populate validateRes if present; we populate in the unhappy path below
	if validateRes != nil {
		requestID = validateRes.RequestID
		assertionID = validateRes.AssertionID
		assertion = validateRes.Assertion
	}

	// note: if err is a saml.ValidateError, then this method continues to flow
	// through; we need to log such errors, which requires much of the same code
	// as the happy path
	if err != nil {
		slog.InfoContext(ctx, "acs_validate_err", "err", err)
		validateFailed = true

		var validateError *saml.ValidateError
		if errors.As(err, &validateError) {
			requestID = validateError.RequestID
			assertionID = validateError.AssertionID
			assertion = validateError.Assertion
			malformedAssertion = validateError.MalformedAssertion
			unsignedAssertion = validateError.UnsignedAssertion
			expiredAssertion = validateError.ExpiredAssertion
			badIssuer = validateError.BadIDPEntityID
			badAudience = validateError.BadSPEntityID
			badSignatureAlgorithm = validateError.BadSignatureAlgorithm
			badDigestAlgorithm = validateError.BadDigestAlgorithm
			badCertificate = validateError.BadCertificate
		} else {
			panic(err)
		}
	}

	if malformedAssertion {
		http.Error(w, "malformed assertion", http.StatusBadRequest)
		return
	}

	if expiredAssertion {
		http.Error(w, "expired assertion", http.StatusBadRequest)
		return
	}

	alreadyProcessed, err := s.Store.AuthCheckAssertionAlreadyProcessed(ctx, requestID)
	if err != nil {
		if errors.Is(err, store.InvalidSAMLRequestID) {
			http.Error(w, "invalid saml request id", http.StatusBadRequest)
			return
		}
		panic(err)
	}

	if alreadyProcessed {
		http.Error(w, "assertion previously processed", http.StatusBadRequest)
		return
	}

	// data that's only possible to validate if saml.Validate didn't return an
	// error
	var (
		badSubjectID        *string
		email               string
		domainMismatchEmail *string
	)

	if validateRes != nil {
		// extract a domain from the subject ID
		subjectEmailDomain, err := emailaddr.Parse(validateRes.SubjectID)
		if err != nil {
			// the subject ID isn't an email
			badSubjectID = &validateRes.SubjectID
		}

		// the subject ID is an email, validate that it's from the set of
		// organization domains
		if badSubjectID == nil {
			email = validateRes.SubjectID

			var domainOk bool
			for _, domain := range dataRes.OrganizationDomains {
				if domain == subjectEmailDomain {
					domainOk = true
				}
			}
			if !domainOk {
				domainMismatchEmail = &subjectEmailDomain
			}
		}
	}

	var subjectIDPAttributes map[string]string
	if validateRes != nil {
		subjectIDPAttributes = validateRes.SubjectAttributes
	}

	createSAMLLoginRes, err := s.Store.AuthUpsertReceiveAssertionData(ctx, &store.AuthUpsertSAMLLoginEventRequest{
		SAMLConnectionID:                     samlConnID,
		SAMLAssertionID:                      &assertionID,
		SAMLFlowID:                           requestID,
		Email:                                email,
		SubjectIDPAttributes:                 subjectIDPAttributes,
		SAMLAssertion:                        assertion,
		ErrorUnsignedAssertion:               unsignedAssertion,
		ErrorBadIssuer:                       badIssuer,
		ErrorBadAudience:                     badAudience,
		ErrorBadSignatureAlgorithm:           badSignatureAlgorithm,
		ErrorBadDigestAlgorithm:              badDigestAlgorithm,
		ErrorBadCertificate:                  badCertificate,
		ErrorBadSubjectID:                    badSubjectID,
		ErrorEmailOutsideOrganizationDomains: domainMismatchEmail,
	})
	if err != nil {
		if errors.Is(err, store.ErrDuplicateAssertionID) {
			http.Error(w, "assertion previously processed", http.StatusBadRequest)
			return
		}

		if errors.Is(err, store.ErrSAMLConnectionIDMismatch) {
			http.Error(w, "assertion not intended for this SAML connection", http.StatusBadRequest)
			return
		}

		panic(err)
	}

	// present an error to the end user depending on their settings
	if unsignedAssertion {
		if err := errorTemplate.Execute(w, &errorTemplateData{
			ErrorMessage: "SAML assertion is unsigned. This needs to be fixed in the Identity Provider.",
			SAMLFlowID:   createSAMLLoginRes.SAMLFlowID,
		}); err != nil {
			panic(fmt.Errorf("acsTemplate.Execute: %w", err))
		}
		return
	}
	if badIssuer != nil {
		if err := errorTemplate.Execute(w, &errorTemplateData{
			ErrorMessage:    "Incorrect IDP Entity ID. This needs to be fixed in the Service Provider.",
			SAMLFlowID:      createSAMLLoginRes.SAMLFlowID,
			WantIDPEntityID: dataRes.IDPEntityID,
			GotIDPEntityID:  *badIssuer,
		}); err != nil {
			panic(fmt.Errorf("acsTemplate.Execute: %w", err))
		}
		return
	}
	if badAudience != nil {
		if err := errorTemplate.Execute(w, &errorTemplateData{
			ErrorMessage:            "Incorrect SP Entity ID. This needs to be fixed in the Identity Provider.",
			SAMLFlowID:              createSAMLLoginRes.SAMLFlowID,
			WantAudienceRestriction: dataRes.SPEntityID,
			GotAudienceRestriction:  *badAudience,
		}); err != nil {
			panic(fmt.Errorf("acsTemplate.Execute: %w", err))
		}
		return
	}
	if badSignatureAlgorithm != nil {
		if err := errorTemplate.Execute(w, &errorTemplateData{
			ErrorMessage:          "Incorrect signature algorithm. This needs to be fixed in the Identity Provider.",
			SAMLFlowID:            createSAMLLoginRes.SAMLFlowID,
			GotSignatureAlgorithm: *badSignatureAlgorithm,
		}); err != nil {
			panic(fmt.Errorf("acsTemplate.Execute: %w", err))
		}
		return
	}
	if badDigestAlgorithm != nil {
		if err := errorTemplate.Execute(w, &errorTemplateData{
			ErrorMessage:       "Incorrect digest algorithm. This needs to be fixed in the Identity Provider.",
			SAMLFlowID:         createSAMLLoginRes.SAMLFlowID,
			GotDigestAlgorithm: *badDigestAlgorithm,
		}); err != nil {
			panic(fmt.Errorf("acsTemplate.Execute: %w", err))
		}
		return
	}
	if badCertificate != nil {
		wantCertPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})

		gotCertPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: badCertificate.Raw,
		})

		if err := errorTemplate.Execute(w, &errorTemplateData{
			ErrorMessage:       "Incorrect certificate. This needs to be fixed in the Service Provider.",
			SAMLFlowID:         createSAMLLoginRes.SAMLFlowID,
			WantCertificatePEM: string(wantCertPEM),
			GotCertificatePEM:  string(gotCertPEM),
		}); err != nil {
			panic(fmt.Errorf("acsTemplate.Execute: %w", err))
		}
		return
	}
	if badSubjectID != nil {
		if err := errorTemplate.Execute(w, &errorTemplateData{
			ErrorMessage: "Subject ID must be an email address. This needs to be fixed in the Identity Provider.",
			SAMLFlowID:   createSAMLLoginRes.SAMLFlowID,
			GotSubjectID: *badSubjectID,
		}); err != nil {
			panic(fmt.Errorf("acsTemplate.Execute: %w", err))
		}
		return
	}
	if domainMismatchEmail != nil {
		if err := errorTemplate.Execute(w, &errorTemplateData{
			ErrorMessage:     "Subject ID email address is not from the list of allowed domains. This needs to be fixed in the Identity Provider.",
			SAMLFlowID:       createSAMLLoginRes.SAMLFlowID,
			GotSubjectID:     validateRes.SubjectID,
			WantEmailDomains: strings.Join(dataRes.OrganizationDomains, ", "),
		}); err != nil {
			panic(fmt.Errorf("acsTemplate.Execute: %w", err))
		}
		return
	}

	// past this point, we will presume the request is valid; panic to ensure we haven't missed problems
	if validateFailed {
		panic("acs_validate_failed")
	}

	// if the saml flow was created as part of the oauth-style flow, then redirect in the OAuth way
	if createSAMLLoginRes.SAMLFlowIsOAuth {
		redirectURL, err := url.Parse(dataRes.EnvironmentOAuthRedirectURI)
		if err != nil {
			panic(err)
		}

		redirectQuery := url.Values{}
		redirectQuery.Set("code", createSAMLLoginRes.Token)
		redirectQuery.Set("state", createSAMLLoginRes.State)
		redirectURL.RawQuery = redirectQuery.Encode()
		redirect := redirectURL.String()

		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}

	// if the saml flow was created as part of the admin test mode, then
	// redirect to the admin ui
	if createSAMLLoginRes.SAMLFlowTestModeIDP != "" {
		redirectURL, err := url.Parse(dataRes.EnvironmentAdminTestModeURL)
		if err != nil {
			panic(err)
		}

		attributes, err := json.Marshal(validateRes.SubjectAttributes)
		if err != nil {
			panic(err)
		}

		redirectQuery := url.Values{}
		redirectQuery.Set("idp", createSAMLLoginRes.SAMLFlowTestModeIDP)
		redirectQuery.Set("saml_connection_id", samlConnID)
		redirectQuery.Set("email", email)
		redirectQuery.Set("attributes", string(attributes))
		redirectURL.RawQuery = redirectQuery.Encode()
		redirect := redirectURL.String()

		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}

	redirectURL, err := url.Parse(dataRes.EnvironmentRedirectURL)
	if err != nil {
		panic(err)
	}

	// this log must happen before we put the sensitive the saml access code
	// into the url
	slog.InfoContext(ctx, "redirect_saml_acs_success", "redirect_url", redirectURL.String())

	redirectQuery := url.Values{}
	redirectQuery.Set("saml_access_code", createSAMLLoginRes.Token)
	redirectURL.RawQuery = redirectQuery.Encode()
	redirect := redirectURL.String()

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}
