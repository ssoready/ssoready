package saml

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
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
	BadIssuer         *string
	BadAudience       *string
	RequestID         string
	Assertion         string
	SubjectID         string
	SubjectAttributes map[string]string
}

func (r *ValidateResponse) IsValid() bool {
	return r.BadIssuer == nil && r.BadAudience == nil
}

var (
	ErrExpired = fmt.Errorf("saml response expired")
)

func Validate(req *ValidateRequest) (*ValidateResponse, error) {
	data, err := base64.StdEncoding.DecodeString(req.SAMLResponse)
	if err != nil {
		return nil, fmt.Errorf("parse saml response: %w", err)
	}

	var samlRes samltypes.Response
	if err := xml.Unmarshal(data, &samlRes); err != nil {
		return nil, fmt.Errorf("unmarshal saml response: %w", err)
	}

	attrs := map[string]string{}
	for _, attr := range samlRes.Assertion.AttributeStatement.Attributes {
		attrs[attr.Name] = attr.Value
	}

	res := ValidateResponse{
		RequestID:         samlRes.Assertion.Subject.SubjectConfirmation.SubjectConfirmationData.InResponseTo,
		Assertion:         string(data),
		SubjectID:         samlRes.Assertion.Subject.NameID.Value,
		SubjectAttributes: attrs,
	}

	if err := dsig.Verify(req.IDPCertificate, data); err != nil {
		return &res, fmt.Errorf("verify signature: %w", err)
	}

	if samlRes.Assertion.Issuer.Name != req.IDPEntityID {
		res.BadIssuer = &samlRes.Assertion.Issuer.Name
		return &res, nil
	}

	if samlRes.Assertion.Conditions.AudienceRestriction.Audience.Name != req.SPEntityID {
		res.BadAudience = &samlRes.Assertion.Conditions.AudienceRestriction.Audience.Name
		return &res, nil
	}

	if req.Now.Before(samlRes.Assertion.Conditions.NotBefore) {
		return &res, ErrExpired
	}

	if req.Now.After(samlRes.Assertion.Conditions.NotOnOrAfter) {
		return &res, ErrExpired
	}

	return &res, nil
}
