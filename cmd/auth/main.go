package main

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssoready/ssoready/internal/saml"
	"github.com/ssoready/ssoready/internal/store"
)

func main() {
	db, err := pgxpool.New(context.Background(), "postgres://postgres:password@localhost/postgres")
	if err != nil {
		panic(err)
	}

	store_ := store.New(db)

	r := mux.NewRouter()
	r.HandleFunc("/saml/{saml_conn_id}/init", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		samlConnID := mux.Vars(r)["saml_conn_id"]
		getSamlConnRes, err := store_.GetSAMLConnectionByID(ctx, &store.GetSAMLConnectionByIDRequest{ID: samlConnID})
		if err != nil {
			panic(err)
		}

		initRes, err := saml.Init(&saml.InitRequest{
			IDPRedirectURL: *getSamlConnRes.SAMLConnection.IdpRedirectUrl,
			SPEntityID:     fmt.Sprintf("http://localhost:8080/saml/%s", getSamlConnRes.SAMLConnection.ID),
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
		getSamlConnRes, err := store_.GetSAMLConnectionByID(ctx, &store.GetSAMLConnectionByIDRequest{ID: samlConnID})
		if err != nil {
			panic(err)
		}

		if err := r.ParseForm(); err != nil {
			panic(err)
		}

		cert, err := x509.ParseCertificate(getSamlConnRes.SAMLConnection.IdpX509Certificate)
		if err != nil {
			panic(err)
		}

		validateRes, err := saml.Validate(&saml.ValidateRequest{
			SAMLResponse:   r.FormValue("SAMLResponse"),
			IDPCertificate: cert,
			IDPEntityID:    *getSamlConnRes.SAMLConnection.IdpEntityID,
			SPEntityID:     fmt.Sprintf("http://localhost:8080/saml/%s", getSamlConnRes.SAMLConnection.ID),
			Now:            time.Now(),
		})
		if err != nil {
			panic(err)
		}

		if err := json.NewEncoder(w).Encode(validateRes); err != nil {
			panic(err)
		}
	}).Methods("POST")

	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		panic(err)
	}
}
