package saml

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/ssoready/ssoready/internal/samltypes"
)

type ParseMetadataResponse struct {
	IDPEntityID    string
	IDPCertificate *x509.Certificate
	RedirectURL    string
}

func ParseMetadata(b []byte) (*ParseMetadataResponse, error) {
	var metadata samltypes.Metadata
	if err := xml.Unmarshal(b, &metadata); err != nil {
		return nil, err
	}

	asn1Base64 := metadata.IDPSSODescriptor.KeyDescriptor.KeyInfo.X509Data.X509Certificate.Value
	asn1Base64 = strings.ReplaceAll(asn1Base64, " ", "")
	asn1Base64 = strings.ReplaceAll(asn1Base64, "\n", "")
	asn1Data, err := base64.StdEncoding.DecodeString(asn1Base64)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(asn1Data)
	if err != nil {
		return nil, err
	}

	for _, s := range metadata.IDPSSODescriptor.SingleSignOnServices {
		if s.Binding == "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" {
			return &ParseMetadataResponse{
				IDPEntityID:    metadata.EntityID,
				IDPCertificate: cert,
				RedirectURL:    s.Location,
			}, nil
		}
	}

	return nil, fmt.Errorf("metadata has no HTTP-POST binding")
}
