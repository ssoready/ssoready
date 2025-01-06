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
