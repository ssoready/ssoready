package saml

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/ssoready/ssoready/internal/dsig"
	"github.com/ssoready/ssoready/internal/samlres"
)

type ValidateRequest struct {
	SAMLResponse   string
	IDPCertificate *x509.Certificate
	IDPEntityID    string
	SPEntityID     string
	Now            time.Time
}

type ValidateResponse struct {
	SubjectID         string
	SubjectAttributes map[string]string
}

var (
	ErrBadIssuer   = fmt.Errorf("bad saml response issuer")
	ErrBadAudience = fmt.Errorf("bad saml response audience restriction")
	ErrExpired     = fmt.Errorf("saml response expired")
)

func Validate(req *ValidateRequest) (*ValidateResponse, error) {
	data, err := base64.StdEncoding.DecodeString(req.SAMLResponse)
	if err != nil {
		return nil, fmt.Errorf("parse saml response: %w", err)
	}

	var res samlres.SAMLResponse
	if err := xml.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("unmarshal saml response: %w", err)
	}

	if err := dsig.Verify(req.IDPCertificate, data); err != nil {
		return nil, err
	}

	if res.Assertion.Issuer.Name != req.IDPEntityID {
		return nil, ErrBadIssuer
	}

	if res.Assertion.Conditions.AudienceRestriction.Audience.Name != req.SPEntityID {
		return nil, ErrBadAudience
	}

	if req.Now.Before(res.Assertion.Conditions.NotBefore) {
		return nil, ErrExpired
	}

	if req.Now.After(res.Assertion.Conditions.NotOnOrAfter) {
		return nil, ErrExpired
	}

	attrs := map[string]string{}
	for _, attr := range res.Assertion.AttributeStatement.Attributes {
		attrs[attr.Name] = attr.Value
	}

	return &ValidateResponse{
		SubjectID:         res.Assertion.Subject.NameID.Value,
		SubjectAttributes: attrs,
	}, nil
}
