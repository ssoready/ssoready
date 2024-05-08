package main

import (
	"context"
	"crypto/x509"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssoready/ssoready/internal/emailaddr"
	"github.com/ssoready/ssoready/internal/pagetoken"
	"github.com/ssoready/ssoready/internal/saml"
	"github.com/ssoready/ssoready/internal/store"
)

func main() {
	db, err := pgxpool.New(context.Background(), "postgres://postgres:password@localhost/postgres")
	if err != nil {
		panic(err)
	}

	// todo populate from env
	store_ := store.New(store.NewStoreParams{
		DB:                   db,
		PageEncoder:          pagetoken.Encoder{Secret: [32]byte{}},
		GlobalDefaultAuthURL: "http://localhost:8080",
		SAMLStateSigningKey:  [32]byte{},
	})

	r := mux.NewRouter()
	r.HandleFunc("/saml/{saml_conn_id}/init", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		samlConnID := mux.Vars(r)["saml_conn_id"]
		state := r.URL.Query().Get("state")

		dataRes, err := store_.AuthGetInitData(ctx, &store.AuthGetInitDataRequest{
			SAMLConnectionID: samlConnID,
			State:            state,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		initRes := saml.Init(&saml.InitRequest{
			RequestID:      dataRes.RequestID,
			IDPRedirectURL: dataRes.IDPRedirectURL,
			SPEntityID:     dataRes.SPEntityID,
			RelayState:     state,
		})

		if err := store_.AuthUpsertInitiateData(ctx, &store.AuthUpsertInitiateDataRequest{
			State:           state,
			InitiateRequest: initRes.InitiateRequest,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, initRes.URL, http.StatusSeeOther)
	}).Methods("GET")

	r.HandleFunc("/saml/{saml_conn_id}/acs", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		samlConnID := mux.Vars(r)["saml_conn_id"]

		dataRes, err := store_.AuthGetValidateData(ctx, &store.AuthGetValidateDataRequest{
			SAMLConnectionID: samlConnID,
		})

		if err := r.ParseForm(); err != nil {
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
		if err != nil {
			panic(err)
		}

		// check that the subject IDP ID is an email, and that email belongs to one of the org's domains
		subjectEmailDomain, err := emailaddr.Parse(validateRes.SubjectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var domainOk bool
		for _, domain := range dataRes.OrganizationDomains {
			if domain == subjectEmailDomain {
				domainOk = true
			}
		}

		if !domainOk {
			http.Error(w, "unauthorized subject email address domain", http.StatusBadRequest)
			return
		}

		createSAMLLoginRes, err := store_.AuthUpsertReceiveAssertionData(ctx, &store.AuthUpsertSAMLLoginEventRequest{
			SAMLConnectionID:     samlConnID,
			SubjectID:            validateRes.SubjectID,
			SubjectIDPAttributes: validateRes.SubjectAttributes,
			SAMLFlowID:           validateRes.RequestID,
			SAMLAssertion:        validateRes.Assertion,
		})
		if err != nil {
			panic(err)
		}

		redirectURL, err := url.Parse(dataRes.EnvironmentRedirectURL)
		if err != nil {
			panic(err)
		}

		redirectQuery := url.Values{}
		redirectQuery.Set("access_token", createSAMLLoginRes.Token)
		redirectURL.RawQuery = redirectQuery.Encode()
		redirect := redirectURL.String()

		if err := store_.AuthUpdateAppRedirectURL(ctx, &store.AuthUpdateAppRedirectURLRequest{
			SAMLFlowID:     createSAMLLoginRes.SAMLFlowID,
			AppRedirectURL: redirect,
		}); err != nil {
			panic(err)
		}

		http.Redirect(w, r, redirect, http.StatusSeeOther)
	}).Methods("POST")

	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		panic(err)
	}
}
