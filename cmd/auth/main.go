package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssoready/ssoready/internal/pagetoken"
	"github.com/ssoready/ssoready/internal/saml"
	"github.com/ssoready/ssoready/internal/store"
)

func main() {
	db, err := pgxpool.New(context.Background(), "postgres://postgres:password@localhost/postgres")
	if err != nil {
		panic(err)
	}

	store_ := store.New(db, pagetoken.Encoder{Secret: [32]byte{}}) // todo populate from env

	r := mux.NewRouter()
	r.HandleFunc("/saml/{saml_conn_id}/init", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		samlConnID := mux.Vars(r)["saml_conn_id"]
		samlConn, err := store_.GetSAMLConnectionByID(ctx, &store.GetSAMLConnectionByIDRequest{ID: samlConnID})
		if err != nil {
			panic(err)
		}

		initRes, err := saml.Init(&saml.InitRequest{
			IDPRedirectURL: samlConn.IdpRedirectUrl,
			SPEntityID:     fmt.Sprintf("http://localhost:8080/saml/%s", samlConn.Id),
			RelayState:     "this is a relay state",
		})
		if err != nil {
			panic(err)
		}
		http.Redirect(w, r, initRes.URL, http.StatusSeeOther)
	}).Methods("GET")

	r.HandleFunc("/saml/{saml_conn_id}/acs", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		samlConnID := mux.Vars(r)["saml_conn_id"]
		samlConn, err := store_.GetSAMLConnectionByID(ctx, &store.GetSAMLConnectionByIDRequest{ID: samlConnID})
		if err != nil {
			panic(err)
		}

		org, err := store_.GetOrganizationByID(ctx, &store.GetOrganizationByIDRequest{
			ID: samlConn.OrganizationId,
		})
		if err != nil {
			panic(err)
		}

		env, err := store_.GetEnvironmentByID(ctx, &store.GetEnvironmentByIDRequest{ID: org.EnvironmentId})
		if err != nil {
			panic(err)
		}

		if err := r.ParseForm(); err != nil {
			panic(err)
		}

		cert, err := x509.ParseCertificate(samlConn.IdpX509Certificate)
		if err != nil {
			panic(err)
		}

		validateRes, err := saml.Validate(&saml.ValidateRequest{
			SAMLResponse:   r.FormValue("SAMLResponse"),
			IDPCertificate: cert,
			IDPEntityID:    samlConn.IdpEntityId,
			SPEntityID:     fmt.Sprintf("http://localhost:8080/saml/%s", samlConn.Id),
			Now:            time.Now(),
		})
		if err != nil {
			panic(err)
		}

		createSAMLSessRes, err := store_.CreateSAMLSession(ctx, &store.CreateSAMLSessionRequest{
			SAMLConnectionID:     samlConn.Id,
			SubjectID:            validateRes.SubjectID,
			SubjectIDPAttributes: validateRes.SubjectAttributes,
		})
		if err != nil {
			panic(err)
		}

		redirectURL, err := url.Parse(env.RedirectUrl)
		if err != nil {
			panic(err)
		}

		redirectQuery := url.Values{}
		redirectQuery.Set("access_token", createSAMLSessRes.Token)
		redirectURL.RawQuery = redirectQuery.Encode()

		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
	}).Methods("POST")

	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		panic(err)
	}
}
