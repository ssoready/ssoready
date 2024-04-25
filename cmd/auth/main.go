package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ssoready/ssoready/internal/saml"
)

func main() {
	metadata, err := saml.ParseMetadata([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?><md:EntityDescriptor entityID=\"http://www.okta.com/exkdoocxa1VmjpXmX697\" xmlns:md=\"urn:oasis:names:tc:SAML:2.0:metadata\"><md:IDPSSODescriptor WantAuthnRequestsSigned=\"false\" protocolSupportEnumeration=\"urn:oasis:names:tc:SAML:2.0:protocol\"><md:KeyDescriptor use=\"signing\"><ds:KeyInfo xmlns:ds=\"http://www.w3.org/2000/09/xmldsig#\"><ds:X509Data><ds:X509Certificate>MIIDqjCCApKgAwIBAgIGAY8W9FSqMA0GCSqGSIb3DQEBCwUAMIGVMQswCQYDVQQGEwJVUzETMBEG\n    A1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwNU2FuIEZyYW5jaXNjbzENMAsGA1UECgwET2t0YTEU\n    MBIGA1UECwwLU1NPUHJvdmlkZXIxFjAUBgNVBAMMDXRyaWFsLTEwMjI4NjMxHDAaBgkqhkiG9w0B\n    CQEWDWluZm9Ab2t0YS5jb20wHhcNMjQwNDI1MjAzMDAyWhcNMzQwNDI1MjAzMTAyWjCBlTELMAkG\n    A1UEBhMCVVMxEzARBgNVBAgMCkNhbGlmb3JuaWExFjAUBgNVBAcMDVNhbiBGcmFuY2lzY28xDTAL\n    BgNVBAoMBE9rdGExFDASBgNVBAsMC1NTT1Byb3ZpZGVyMRYwFAYDVQQDDA10cmlhbC0xMDIyODYz\n    MRwwGgYJKoZIhvcNAQkBFg1pbmZvQG9rdGEuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB\n    CgKCAQEAh8g24a5HDZpwtWuA/HP1JuecGMZ1Wh8R3QC/DQb4aNJtNwJlzMN746MQhkEtXI4TYTah\n    3bpbJc5jUFunjZdy8I4+pHCa4wS7lf9Z3c2Ptc9R1XzAX9zhC1Cuj01L69vAinNF8JR1tTx1A7im\n    pAWqjtKEQAZNsWrjo0TkQVZlU2wY/CLW+w/zRmHmxSzuCHIVtD9SkgPVXr/Wr2X2SFUc0miGc09x\n    FKSl1ARIRVf7jrI0hcSpB5lOd4jrZaM6pvYPTHZYsvtvE9IJUtRlD3OAenBeiHBvkzPwbnhIFUm0\n    2Rq9Q7Fvr2CMD8+w/vdgFECelHS0euNVx3uOGydnUh9WOQIDAQABMA0GCSqGSIb3DQEBCwUAA4IB\n    AQBqUvihKyejxTpV/mcm7KQu4g3NUx5blTa1jRj2jCDfbn3YckqGI9i0j8BAHNaZw56Nu7OIzDrL\n    nxsi8uMmdRAJqAQA7iILGAEJuMvHfv2SJkcu2goB9Xl69Kh34UgZd3tucDEgM3cwhUlltU8yV+P2\n    +uzhNaHJkDargKeEI1NQG0lvcFJHP5ESTR9idIipJDdBcSxais3wLkRlhvufp3Rr71Z6TylTVvc3\n    QwAjCyTmfR2YjhQkVVfWdOEwqOYhyIn2d+gUex0gEGOZqzmMgCD20mNkiL+YTEsz5XqDaUDQsLrS\n    whMgwbzHoz7vrWZiwq2K2AYIu8Uh//DZxsDM9g0B</ds:X509Certificate></ds:X509Data></ds:KeyInfo></md:KeyDescriptor><md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</md:NameIDFormat><md:SingleSignOnService Binding=\"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST\" Location=\"https://trial-1022863.okta.com/app/trial-1022863_oktalocalhostbis_1/exkdoocxa1VmjpXmX697/sso/saml\"/><md:SingleSignOnService Binding=\"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect\" Location=\"https://trial-1022863.okta.com/app/trial-1022863_oktalocalhostbis_1/exkdoocxa1VmjpXmX697/sso/saml\"/></md:IDPSSODescriptor></md:EntityDescriptor>\n"))
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/init", func(w http.ResponseWriter, r *http.Request) {
		initRes, err := saml.Init(&saml.InitRequest{
			IDPRedirectURL: "https://trial-1022863.okta.com/app/trial-1022863_oktalocalhostbis_1/exkdoocxa1VmjpXmX697/sso/saml",
			SPEntityID:     "http://localhost:8080",
			RelayState:     "this is a relay state",
		})
		if err != nil {
			panic(err)
		}
		http.Redirect(w, r, initRes.URL, http.StatusSeeOther)
	}).Methods("GET")

	r.HandleFunc("/acs", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			panic(err)
		}

		validateRes, err := saml.Validate(&saml.ValidateRequest{
			SAMLResponse:   r.FormValue("SAMLResponse"),
			IDPCertificate: metadata.IDPCertificate,
			IDPEntityID:    metadata.IDPEntityID,
			SPEntityID:     "http://localhost:8080",
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
