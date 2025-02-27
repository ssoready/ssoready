package saml_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ssoready/ssoready/internal/saml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_KnownGoodAssertions(t *testing.T) {
	entries, err := os.ReadDir("testdata/assertions")
	assert.NoError(t, err)

	for _, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			_, err := validateFromDir(fmt.Sprintf("testdata/assertions/%s", entry.Name()))
			assert.NoError(t, err)
		})
	}
}

func TestValidate_GoodAssertionData(t *testing.T) {
	res, err := validateFromDir("testdata/assertions/okta")
	require.NoError(t, err)

	assert.Equal(t, "<?xml version=\"1.0\" encoding=\"UTF-8\"?><saml2p:Response Destination=\"http://localhost:8080\" ID=\"id35528194005172931133953195\" IssueInstant=\"2024-04-25T20:31:55.494Z\" Version=\"2.0\" xmlns:saml2p=\"urn:oasis:names:tc:SAML:2.0:protocol\"><saml2:Issuer Format=\"urn:oasis:names:tc:SAML:2.0:nameid-format:entity\" xmlns:saml2=\"urn:oasis:names:tc:SAML:2.0:assertion\">http://www.okta.com/exkdoocxa1VmjpXmX697</saml2:Issuer><ds:Signature xmlns:ds=\"http://www.w3.org/2000/09/xmldsig#\"><ds:SignedInfo><ds:CanonicalizationMethod Algorithm=\"http://www.w3.org/2001/10/xml-exc-c14n#\"/><ds:SignatureMethod Algorithm=\"http://www.w3.org/2001/04/xmldsig-more#rsa-sha256\"/><ds:Reference URI=\"#id35528194005172931133953195\"><ds:Transforms><ds:Transform Algorithm=\"http://www.w3.org/2000/09/xmldsig#enveloped-signature\"/><ds:Transform Algorithm=\"http://www.w3.org/2001/10/xml-exc-c14n#\"/></ds:Transforms><ds:DigestMethod Algorithm=\"http://www.w3.org/2001/04/xmlenc#sha256\"/><ds:DigestValue>tQ3cGy9Kax5v8DdRTNTVPboMtL5viRVZLNmBIgpx/rQ=</ds:DigestValue></ds:Reference></ds:SignedInfo><ds:SignatureValue>Jbtjo4MLglMSc6SopDHj2ZdRf8IA0bT5nlLeaysYgGlj0kd3gO6vYFzsybD6EqRiZvrUrOJU8JANuz17vpPxSGLmt8h1N1Uy0vVRpL3VQYU7KNgr6o2xtSU87IzBKCaGfFqPqN4CLaCs1wbKkAdkxKnwdEo6kHE//hAEckDofmKXdEJDihy8h6uUxO/EwKJgg9+G/8UYD3YiKpeFHfJTI0W+rDKLGmPXbRvHNF/JriltOTPSSZ8noQk2fz7WWYyO0F179MDMBDyxRHhA1uOf9JCYr28pCQ9iPQIIQnABVgAdaq++hixIHhvR4jNrwpGItwJb7aqCqd28TuXXzBUkxw==</ds:SignatureValue><ds:KeyInfo><ds:X509Data><ds:X509Certificate>MIIDqjCCApKgAwIBAgIGAY8W9FSqMA0GCSqGSIb3DQEBCwUAMIGVMQswCQYDVQQGEwJVUzETMBEG\n    A1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwNU2FuIEZyYW5jaXNjbzENMAsGA1UECgwET2t0YTEU\n    MBIGA1UECwwLU1NPUHJvdmlkZXIxFjAUBgNVBAMMDXRyaWFsLTEwMjI4NjMxHDAaBgkqhkiG9w0B\n    CQEWDWluZm9Ab2t0YS5jb20wHhcNMjQwNDI1MjAzMDAyWhcNMzQwNDI1MjAzMTAyWjCBlTELMAkG\n    A1UEBhMCVVMxEzARBgNVBAgMCkNhbGlmb3JuaWExFjAUBgNVBAcMDVNhbiBGcmFuY2lzY28xDTAL\n    BgNVBAoMBE9rdGExFDASBgNVBAsMC1NTT1Byb3ZpZGVyMRYwFAYDVQQDDA10cmlhbC0xMDIyODYz\n    MRwwGgYJKoZIhvcNAQkBFg1pbmZvQG9rdGEuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB\n    CgKCAQEAh8g24a5HDZpwtWuA/HP1JuecGMZ1Wh8R3QC/DQb4aNJtNwJlzMN746MQhkEtXI4TYTah\n    3bpbJc5jUFunjZdy8I4+pHCa4wS7lf9Z3c2Ptc9R1XzAX9zhC1Cuj01L69vAinNF8JR1tTx1A7im\n    pAWqjtKEQAZNsWrjo0TkQVZlU2wY/CLW+w/zRmHmxSzuCHIVtD9SkgPVXr/Wr2X2SFUc0miGc09x\n    FKSl1ARIRVf7jrI0hcSpB5lOd4jrZaM6pvYPTHZYsvtvE9IJUtRlD3OAenBeiHBvkzPwbnhIFUm0\n    2Rq9Q7Fvr2CMD8+w/vdgFECelHS0euNVx3uOGydnUh9WOQIDAQABMA0GCSqGSIb3DQEBCwUAA4IB\n    AQBqUvihKyejxTpV/mcm7KQu4g3NUx5blTa1jRj2jCDfbn3YckqGI9i0j8BAHNaZw56Nu7OIzDrL\n    nxsi8uMmdRAJqAQA7iILGAEJuMvHfv2SJkcu2goB9Xl69Kh34UgZd3tucDEgM3cwhUlltU8yV+P2\n    +uzhNaHJkDargKeEI1NQG0lvcFJHP5ESTR9idIipJDdBcSxais3wLkRlhvufp3Rr71Z6TylTVvc3\n    QwAjCyTmfR2YjhQkVVfWdOEwqOYhyIn2d+gUex0gEGOZqzmMgCD20mNkiL+YTEsz5XqDaUDQsLrS\n    whMgwbzHoz7vrWZiwq2K2AYIu8Uh//DZxsDM9g0B</ds:X509Certificate></ds:X509Data></ds:KeyInfo></ds:Signature><saml2p:Status xmlns:saml2p=\"urn:oasis:names:tc:SAML:2.0:protocol\"><saml2p:StatusCode Value=\"urn:oasis:names:tc:SAML:2.0:status:Success\"/></saml2p:Status><saml2:Assertion ID=\"id35528194006743571812188338\" IssueInstant=\"2024-04-25T20:31:55.494Z\" Version=\"2.0\" xmlns:saml2=\"urn:oasis:names:tc:SAML:2.0:assertion\"><saml2:Issuer Format=\"urn:oasis:names:tc:SAML:2.0:nameid-format:entity\" xmlns:saml2=\"urn:oasis:names:tc:SAML:2.0:assertion\">http://www.okta.com/exkdoocxa1VmjpXmX697</saml2:Issuer><ds:Signature xmlns:ds=\"http://www.w3.org/2000/09/xmldsig#\"><ds:SignedInfo><ds:CanonicalizationMethod Algorithm=\"http://www.w3.org/2001/10/xml-exc-c14n#\"/><ds:SignatureMethod Algorithm=\"http://www.w3.org/2001/04/xmldsig-more#rsa-sha256\"/><ds:Reference URI=\"#id35528194006743571812188338\"><ds:Transforms><ds:Transform Algorithm=\"http://www.w3.org/2000/09/xmldsig#enveloped-signature\"/><ds:Transform Algorithm=\"http://www.w3.org/2001/10/xml-exc-c14n#\"/></ds:Transforms><ds:DigestMethod Algorithm=\"http://www.w3.org/2001/04/xmlenc#sha256\"/><ds:DigestValue>gYuidj1kP4bdhylZxf86HtfmknIINpURdJGSpKXoM3I=</ds:DigestValue></ds:Reference></ds:SignedInfo><ds:SignatureValue>IJ6EEwpVGEOCTJX6xOlO43/HtZ9JgOYODlZepGfUHWVYYkKee6LW/zyCgKPPtHObgzbUjjyRjl/0yMQQ1FhB7K8KJp0dYu/1iguHxKJdVSb2fdAerKzpGuJq2WXw1BtKp+UmCfDi1SFjCrXf8noJrwYnpRI+wVgHC+QnSE1S49y+E2FLt5WY18D8KvMj8X7SKDNOhgUTUCErnG3PVH6gR4WRjTv8Ea3C+jn4jKKLSbvNgVJ9WuI8ZMM90g+LyVEWvPaR2zs9SQgAJG5VHAKtS7c8VCINRwyNb+oh2hdpaSMjPrxdPEjdD73YuZttupZGjnPbI4PsS2kg4OL4py0Ozw==</ds:SignatureValue><ds:KeyInfo><ds:X509Data><ds:X509Certificate>MIIDqjCCApKgAwIBAgIGAY8W9FSqMA0GCSqGSIb3DQEBCwUAMIGVMQswCQYDVQQGEwJVUzETMBEG\n    A1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwNU2FuIEZyYW5jaXNjbzENMAsGA1UECgwET2t0YTEU\n    MBIGA1UECwwLU1NPUHJvdmlkZXIxFjAUBgNVBAMMDXRyaWFsLTEwMjI4NjMxHDAaBgkqhkiG9w0B\n    CQEWDWluZm9Ab2t0YS5jb20wHhcNMjQwNDI1MjAzMDAyWhcNMzQwNDI1MjAzMTAyWjCBlTELMAkG\n    A1UEBhMCVVMxEzARBgNVBAgMCkNhbGlmb3JuaWExFjAUBgNVBAcMDVNhbiBGcmFuY2lzY28xDTAL\n    BgNVBAoMBE9rdGExFDASBgNVBAsMC1NTT1Byb3ZpZGVyMRYwFAYDVQQDDA10cmlhbC0xMDIyODYz\n    MRwwGgYJKoZIhvcNAQkBFg1pbmZvQG9rdGEuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB\n    CgKCAQEAh8g24a5HDZpwtWuA/HP1JuecGMZ1Wh8R3QC/DQb4aNJtNwJlzMN746MQhkEtXI4TYTah\n    3bpbJc5jUFunjZdy8I4+pHCa4wS7lf9Z3c2Ptc9R1XzAX9zhC1Cuj01L69vAinNF8JR1tTx1A7im\n    pAWqjtKEQAZNsWrjo0TkQVZlU2wY/CLW+w/zRmHmxSzuCHIVtD9SkgPVXr/Wr2X2SFUc0miGc09x\n    FKSl1ARIRVf7jrI0hcSpB5lOd4jrZaM6pvYPTHZYsvtvE9IJUtRlD3OAenBeiHBvkzPwbnhIFUm0\n    2Rq9Q7Fvr2CMD8+w/vdgFECelHS0euNVx3uOGydnUh9WOQIDAQABMA0GCSqGSIb3DQEBCwUAA4IB\n    AQBqUvihKyejxTpV/mcm7KQu4g3NUx5blTa1jRj2jCDfbn3YckqGI9i0j8BAHNaZw56Nu7OIzDrL\n    nxsi8uMmdRAJqAQA7iILGAEJuMvHfv2SJkcu2goB9Xl69Kh34UgZd3tucDEgM3cwhUlltU8yV+P2\n    +uzhNaHJkDargKeEI1NQG0lvcFJHP5ESTR9idIipJDdBcSxais3wLkRlhvufp3Rr71Z6TylTVvc3\n    QwAjCyTmfR2YjhQkVVfWdOEwqOYhyIn2d+gUex0gEGOZqzmMgCD20mNkiL+YTEsz5XqDaUDQsLrS\n    whMgwbzHoz7vrWZiwq2K2AYIu8Uh//DZxsDM9g0B</ds:X509Certificate></ds:X509Data></ds:KeyInfo></ds:Signature><saml2:Subject xmlns:saml2=\"urn:oasis:names:tc:SAML:2.0:assertion\"><saml2:NameID Format=\"urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress\">ulysse.carion@codomaindata.com</saml2:NameID><saml2:SubjectConfirmation Method=\"urn:oasis:names:tc:SAML:2.0:cm:bearer\"><saml2:SubjectConfirmationData NotOnOrAfter=\"2024-04-25T20:36:55.494Z\" Recipient=\"http://localhost:8080\"/></saml2:SubjectConfirmation></saml2:Subject><saml2:Conditions NotBefore=\"2024-04-25T20:26:55.494Z\" NotOnOrAfter=\"2024-04-25T20:36:55.494Z\" xmlns:saml2=\"urn:oasis:names:tc:SAML:2.0:assertion\"><saml2:AudienceRestriction><saml2:Audience>http://localhost:8080</saml2:Audience></saml2:AudienceRestriction></saml2:Conditions><saml2:AuthnStatement AuthnInstant=\"2024-04-25T20:26:25.134Z\" SessionIndex=\"id1714077115304.328863927\" xmlns:saml2=\"urn:oasis:names:tc:SAML:2.0:assertion\"><saml2:AuthnContext><saml2:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport</saml2:AuthnContextClassRef></saml2:AuthnContext></saml2:AuthnStatement></saml2:Assertion></saml2p:Response>\n", res.Assertion)
	assert.Equal(t, "id35528194006743571812188338", res.AssertionID)
	assert.Equal(t, "", res.RequestID)
	assert.Equal(t, "ulysse.carion@codomaindata.com", res.SubjectID)
	assert.Equal(t, map[string]string{}, res.SubjectAttributes)
}

func TestValidate_UnsignedAssertion(t *testing.T) {
	// modified from okta, but assertion.xml has the signature stripped
	_, err := validateFromDir("testdata/bad-assertions/unsigned-assertion")
	var validateError *saml.ValidateError
	if !errors.As(err, &validateError) {
		t.Fatalf("bad error: %v", err)
	}

	require.True(t, validateError.UnsignedAssertion)
}

func TestValidate_BadIDPEntityID(t *testing.T) {
	// modified from okta, but metadata.xml has a different idp entity id
	_, err := validateFromDir("testdata/bad-assertions/bad-idp-entity-id")
	var validateError *saml.ValidateError
	if !errors.As(err, &validateError) {
		t.Fatalf("bad error: %v", err)
	}

	require.NotEmpty(t, validateError.BadIDPEntityID)
	assert.Equal(t, "http://www.okta.com/exkdoocxa1VmjpXmX697", *validateError.BadIDPEntityID)
}

func TestValidate_BadSPEntityID(t *testing.T) {
	// modified from okta, but params.json has a different sp entity id
	_, err := validateFromDir("testdata/bad-assertions/bad-sp-entity-id")
	var validateError *saml.ValidateError
	if !errors.As(err, &validateError) {
		t.Fatalf("bad error: %v", err)
	}

	require.NotEmpty(t, validateError.BadSPEntityID)
	assert.Equal(t, "http://localhost:8080", *validateError.BadSPEntityID)
}

func TestValidate_BadSignatureAlgorithm(t *testing.T) {
	// modified from okta, but assertion.xml has a modified signature algorithm
	_, err := validateFromDir("testdata/bad-assertions/bad-signature-algorithm")
	var validateError *saml.ValidateError
	if !errors.As(err, &validateError) {
		t.Fatalf("bad error: %v", err)
	}

	require.NotEmpty(t, validateError.BadSignatureAlgorithm)
	assert.Equal(t, "BAD_SIGNATURE_ALGORITHM", *validateError.BadSignatureAlgorithm)
}

func TestValidate_BadDigestAlgorithm(t *testing.T) {
	// modified from okta, but assertion.xml has a modified digest algorithm
	_, err := validateFromDir("testdata/bad-assertions/bad-digest-algorithm")
	var validateError *saml.ValidateError
	if !errors.As(err, &validateError) {
		t.Fatalf("bad error: %v", err)
	}

	require.NotEmpty(t, validateError.BadDigestAlgorithm)
	assert.Equal(t, "BAD_DIGEST_ALGORITHM", *validateError.BadDigestAlgorithm)
}

func TestValidate_BadCertificate(t *testing.T) {
	// modified from okta, but metadata.xml has a modified certificate
	_, err := validateFromDir("testdata/bad-assertions/bad-certificate")
	var validateError *saml.ValidateError
	if !errors.As(err, &validateError) {
		t.Fatalf("bad error: %v", err)
	}

	require.NotEmpty(t, validateError.BadCertificate)
}

func TestValidate_NoCertificate(t *testing.T) {
	// modified from okta, but no KeyInfo
	_, err := validateFromDir("testdata/bad-assertions/no-certificate")
	var validateError *saml.ValidateError
	if !errors.As(err, &validateError) {
		t.Fatalf("bad error: %v", err)
	}

	require.NotEmpty(t, validateError.UnsignedAssertion)
}

func TestValidate_BadAssertionUTF8(t *testing.T) {
	// modified from okta, but assertion.xml is just \x00 (and a newline)
	_, err := validateFromDir("testdata/bad-assertions/bad-assertion-utf8")
	var validateError *saml.ValidateError
	if !errors.As(err, &validateError) {
		t.Fatalf("bad error: %v", err)
	}

	require.True(t, validateError.MalformedAssertion)
}

func validateFromDir(path string) (*saml.ValidateResponse, error) {
	assertion, err := os.ReadFile(fmt.Sprintf("%s/assertion.xml", path))
	if err != nil {
		return nil, fmt.Errorf("read assertion: %w", err)
	}

	metadata, err := os.ReadFile(fmt.Sprintf("%s/metadata.xml", path))
	if err != nil {
		return nil, fmt.Errorf("read metadata: %w", err)
	}

	params, err := os.ReadFile(fmt.Sprintf("%s/params.json", path))
	if err != nil {
		return nil, fmt.Errorf("read params: %w", err)
	}

	parseMetadataRes, err := saml.ParseMetadata(metadata)
	if err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}

	var paramData struct {
		SPEntityID string    `json:"sp_entity_id"`
		Now        time.Time `json:"now"`
	}
	if err := json.Unmarshal(params, &paramData); err != nil {
		return nil, fmt.Errorf("unmarshal params: %w", err)
	}

	validateRes, err := saml.Validate(&saml.ValidateRequest{
		SAMLResponse:   base64.StdEncoding.EncodeToString(assertion),
		IDPCertificate: parseMetadataRes.IDPCertificate,
		IDPEntityID:    parseMetadataRes.IDPEntityID,
		SPEntityID:     paramData.SPEntityID,
		Now:            paramData.Now,
	})
	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	return validateRes, nil
}
