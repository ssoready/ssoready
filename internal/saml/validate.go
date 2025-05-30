package saml

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"time"

	"github.com/ssoready/ssoready/internal/saml/dsig"
	"github.com/ssoready/ssoready/internal/saml/samltypes"
)

type ValidateRequest struct {
	SAMLResponse   string
	IDPCertificate *x509.Certificate
	IDPEntityID    string
	SPEntityID     string
	Now            time.Time
}

type ValidateResponse struct {
	RequestID         string
	AssertionID       string
	Assertion         string
	SubjectID         string
	SubjectAttributes map[string]string
}

type ValidateError struct {
	RequestID   string
	AssertionID string
	Assertion   string

	MalformedAssertion    bool
	UnsignedAssertion     bool
	ExpiredAssertion      bool
	BadIDPEntityID        *string
	BadSPEntityID         *string
	BadSignatureAlgorithm *string
	BadDigestAlgorithm    *string
	BadCertificate        *x509.Certificate
}

func (e *ValidateError) Error() string {
	if e.MalformedAssertion {
		return "saml assertion is malformed"
	}

	if e.UnsignedAssertion {
		return "saml assertion is unsigned"
	}

	if e.BadIDPEntityID != nil {
		return "bad idp entity id: " + *e.BadIDPEntityID
	}

	if e.BadSPEntityID != nil {
		return "bad sp entity id: " + *e.BadSPEntityID
	}

	if e.BadSignatureAlgorithm != nil {
		return "bad signature algorithm: " + *e.BadSignatureAlgorithm
	}

	if e.BadDigestAlgorithm != nil {
		return "bad digest algorithm: " + *e.BadDigestAlgorithm
	}

	if e.BadCertificate != nil {
		return "bad assertion certificate"
	}

	panic("unreachable")
}

func Validate(req *ValidateRequest) (*ValidateResponse, error) {
	unverifiedData, err := base64.StdEncoding.DecodeString(req.SAMLResponse)
	if err != nil {
		return nil, fmt.Errorf("parse saml response: %s: %w", err.Error(), &ValidateError{
			MalformedAssertion: true,
		})
	}

	var unverifiedResponse samltypes.Response
	if err := xml.Unmarshal(unverifiedData, &unverifiedResponse); err != nil {
		return nil, fmt.Errorf("parse saml response: %s: %w", err.Error(), &ValidateError{
			MalformedAssertion: true,
		})
	}

	validateError := &ValidateError{
		Assertion: string(unverifiedData),

		// populate these fields on a preliminary basis, so we can report errors
		// back to the user
		RequestID:   unverifiedResponse.Assertion.Subject.SubjectConfirmation.SubjectConfirmationData.InResponseTo,
		AssertionID: unverifiedResponse.Assertion.ID,
	}

	verifiedData, err := dsig.Verify(req.IDPCertificate, unverifiedData)
	if err != nil {
		if errors.Is(err, dsig.ErrUnsigned) {
			validateError.UnsignedAssertion = true
			return nil, validateError
		}

		var badSigAlgError dsig.BadSignatureAlgorithmError
		if errors.As(err, &badSigAlgError) {
			validateError.BadSignatureAlgorithm = &badSigAlgError.BadAlgorithm
			return nil, validateError
		}

		var badDigestAlgError dsig.BadDigestAlgorithmError
		if errors.As(err, &badDigestAlgError) {
			validateError.BadDigestAlgorithm = &badDigestAlgError.BadAlgorithm
			return nil, validateError
		}

		var badCertificateError dsig.BadCertificateError
		if errors.As(err, &badCertificateError) {
			validateError.BadCertificate = badCertificateError.BadCertificate
			return nil, validateError
		}

		return nil, fmt.Errorf("verify signature: %w", err)
	}

	var assertion samltypes.Assertion
	if err := xml.Unmarshal(verifiedData, &assertion); err != nil {
		panic(err)
	}

	attrs := map[string]string{}
	for _, attr := range assertion.AttributeStatement.Attributes {
		attrs[attr.Name] = attr.Value
	}

	res := ValidateResponse{
		Assertion: string(unverifiedData),

		// For purity's sake, when an assertion is considered legitimate, prefer
		// the RequestID and AssertionID from the canonicalized assertion (in
		// the variable assertion) over the initial input (in
		// unverifiedResponse).
		RequestID:         assertion.Subject.SubjectConfirmation.SubjectConfirmationData.InResponseTo,
		AssertionID:       assertion.ID,
		SubjectID:         assertion.Subject.NameID.Value,
		SubjectAttributes: attrs,
	}

	if assertion.Issuer.Name != req.IDPEntityID {
		validateError.BadIDPEntityID = &assertion.Issuer.Name
		return nil, validateError
	}

	if assertion.Conditions.AudienceRestriction.Audience.Name != req.SPEntityID {
		validateError.BadSPEntityID = &assertion.Conditions.AudienceRestriction.Audience.Name
		return nil, validateError
	}

	if req.Now.Before(assertion.Conditions.NotBefore) {
		validateError.ExpiredAssertion = true
		return nil, validateError
	}

	if req.Now.After(assertion.Conditions.NotOnOrAfter) {
		validateError.ExpiredAssertion = true
		return nil, validateError
	}

	return &res, nil
}
