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

type ValidateProblems struct {
	UnsignedAssertion     bool
	BadIDPEntityID        *string
	BadSPEntityID         *string
	BadSignatureAlgorithm *string
	BadDigestAlgorithm    *string
	BadCertificate        *x509.Certificate
}

type ValidateResponse struct {
	RequestID         string
	AssertionID       string
	Assertion         string
	SubjectID         string
	SubjectAttributes map[string]string
}

var (
	errExpired = fmt.Errorf("saml response expired")
)

func Validate(req *ValidateRequest) (*ValidateResponse, *ValidateProblems, error) {
	unverifiedData, err := base64.StdEncoding.DecodeString(req.SAMLResponse)
	if err != nil {
		return nil, nil, fmt.Errorf("parse saml response: %w", err)
	}

	res := ValidateResponse{
		Assertion: string(unverifiedData),
	}

	verifiedData, err := dsig.Verify(req.IDPCertificate, unverifiedData)
	if err != nil {
		if errors.Is(err, dsig.ErrUnsigned) {
			return &res, &ValidateProblems{UnsignedAssertion: true}, nil
		}

		var badSigAlgError dsig.BadSignatureAlgorithmError
		if errors.As(err, &badSigAlgError) {
			return &res, &ValidateProblems{BadSignatureAlgorithm: &badSigAlgError.BadAlgorithm}, nil
		}

		var badDigestAlgError dsig.BadDigestAlgorithmError
		if errors.As(err, &badDigestAlgError) {
			return &res, &ValidateProblems{BadDigestAlgorithm: &badDigestAlgError.BadAlgorithm}, nil
		}

		var badCertificateError dsig.BadCertificateError
		if errors.As(err, &badCertificateError) {
			return &res, &ValidateProblems{BadCertificate: badCertificateError.BadCertificate}, nil
		}

		return &res, nil, fmt.Errorf("verify signature: %w", err)
	}

	var assertion samltypes.Assertion
	if err := xml.Unmarshal(verifiedData, &assertion); err != nil {
		panic(err)
	}

	attrs := map[string]string{}
	for _, attr := range assertion.AttributeStatement.Attributes {
		attrs[attr.Name] = attr.Value
	}

	res.RequestID = assertion.Subject.SubjectConfirmation.SubjectConfirmationData.InResponseTo
	res.AssertionID = assertion.ID
	res.SubjectID = assertion.Subject.NameID.Value
	res.SubjectAttributes = attrs

	if assertion.Issuer.Name != req.IDPEntityID {
		return &res, &ValidateProblems{BadIDPEntityID: &assertion.Issuer.Name}, nil
	}

	if assertion.Conditions.AudienceRestriction.Audience.Name != req.SPEntityID {
		return &res, &ValidateProblems{BadSPEntityID: &assertion.Conditions.AudienceRestriction.Audience.Name}, nil
	}

	if req.Now.Before(assertion.Conditions.NotBefore) {
		return &res, nil, errExpired
	}

	if req.Now.After(assertion.Conditions.NotOnOrAfter) {
		return &res, nil, errExpired
	}

	return &res, nil, nil
}
