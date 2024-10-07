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
	data, err := base64.StdEncoding.DecodeString(req.SAMLResponse)
	if err != nil {
		return nil, nil, fmt.Errorf("parse saml response: %w", err)
	}

	var samlRes samltypes.Response
	if err := xml.Unmarshal(data, &samlRes); err != nil {
		return nil, nil, fmt.Errorf("unmarshal saml response: %w", err)
	}

	attrs := map[string]string{}
	for _, attr := range samlRes.Assertion.AttributeStatement.Attributes {
		attrs[attr.Name] = attr.Value
	}

	res := ValidateResponse{
		RequestID:         samlRes.Assertion.Subject.SubjectConfirmation.SubjectConfirmationData.InResponseTo,
		AssertionID:       samlRes.Assertion.ID,
		Assertion:         string(data),
		SubjectID:         samlRes.Assertion.Subject.NameID.Value,
		SubjectAttributes: attrs,
	}

	if err := dsig.Verify(req.IDPCertificate, data); err != nil {
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

	if samlRes.Assertion.Issuer.Name != req.IDPEntityID {
		return &res, &ValidateProblems{BadIDPEntityID: &samlRes.Assertion.Issuer.Name}, nil
	}

	if samlRes.Assertion.Conditions.AudienceRestriction.Audience.Name != req.SPEntityID {
		return &res, &ValidateProblems{BadSPEntityID: &samlRes.Assertion.Conditions.AudienceRestriction.Audience.Name}, nil
	}

	if req.Now.Before(samlRes.Assertion.Conditions.NotBefore) {
		return &res, nil, errExpired
	}

	if req.Now.After(samlRes.Assertion.Conditions.NotOnOrAfter) {
		return &res, nil, errExpired
	}

	return &res, nil, nil
}
