package main

import (
	"context"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssoready/conf"
	"github.com/ssoready/ssoready/internal/emailaddr"
	"github.com/ssoready/ssoready/internal/pagetoken"
	"github.com/ssoready/ssoready/internal/saml"
	"github.com/ssoready/ssoready/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

	config := struct {
		ServeAddr            string `conf:"serve-addr,noredact"`
		DB                   string `conf:"db"`
		GlobalDefaultAuthURL string `conf:"global-default-auth-url,noredact"`
		PageEncodingSecret   string `conf:"page-encoding-secret"`
		SAMLStateSigningKey  string `conf:"saml-state-signing-key"`
	}{
		ServeAddr:            "localhost:8080",
		DB:                   "postgres://postgres:password@localhost/postgres",
		GlobalDefaultAuthURL: "http://localhost:8080",
	}

	conf.Load(&config)
	slog.Info("config", "config", conf.Redact(config))

	db, err := pgxpool.New(context.Background(), config.DB)
	if err != nil {
		panic(err)
	}

	pageEncodingSecret, err := parseHexKey(config.PageEncodingSecret)
	if err != nil {
		panic(fmt.Errorf("parse page encoding secret: %w", err))
	}

	samlStateSigningKey, err := parseHexKey(config.SAMLStateSigningKey)
	if err != nil {
		panic(fmt.Errorf("parse saml state signing key: %w", err))
	}

	store_ := store.New(store.NewStoreParams{
		DB:                   db,
		PageEncoder:          pagetoken.Encoder{Secret: pageEncodingSecret},
		GlobalDefaultAuthURL: config.GlobalDefaultAuthURL,
		SAMLStateSigningKey:  samlStateSigningKey,
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

		// assess the validity of the response; note that invalid requests may still have a nil err; the problem details
		// are stored in validateRes
		// todo maybe split out validateRes, validateProblems, err as the signature instead?
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

		var badSubjectID *string
		subjectEmailDomain, err := emailaddr.Parse(validateRes.SubjectID)
		if err != nil {
			badSubjectID = &validateRes.SubjectID
		}

		var domainMismatchEmail *string
		if badSubjectID == nil {
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

		createSAMLLoginRes, err := store_.AuthUpsertReceiveAssertionData(ctx, &store.AuthUpsertSAMLLoginEventRequest{
			SAMLConnectionID:                     samlConnID,
			SAMLFlowID:                           validateRes.RequestID,
			SubjectID:                            validateRes.SubjectID,
			SubjectIDPAttributes:                 validateRes.SubjectAttributes,
			SAMLAssertion:                        validateRes.Assertion,
			ErrorBadIssuer:                       validateRes.BadIssuer,
			ErrorBadAudience:                     validateRes.BadAudience,
			ErrorBadSubjectID:                    badSubjectID,
			ErrorEmailOutsideOrganizationDomains: domainMismatchEmail,
		})
		if err != nil {
			panic(err)
		}

		// present an error to the end user depending on their settings
		// todo make this pretty html
		if validateRes.BadIssuer != nil {
			http.Error(w, "bad issuer", http.StatusBadRequest)
			return
		}
		if validateRes.BadAudience != nil {
			http.Error(w, "bad audience", http.StatusBadRequest)
			return
		}
		if badSubjectID != nil {
			http.Error(w, "bad subject id", http.StatusBadRequest)
			return
		}
		if domainMismatchEmail != nil {
			http.Error(w, "bad email domain", http.StatusBadRequest)
			return
		}

		// past this point, we presume the request is valid

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

	r.HandleFunc("/internal/health", func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "health")
		w.WriteHeader(http.StatusOK)
	}).Methods("GET")

	slog.Info("serve")
	if err := http.ListenAndServe(config.ServeAddr, r); err != nil {
		panic(err)
	}
}

func parseHexKey(s string) ([32]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return [32]byte{}, err
	}

	if len(b) > 32 {
		return [32]byte{}, fmt.Errorf("key must encode 32 bytes")
	}

	var k [32]byte
	copy(k[:], b)
	return k, nil
}
