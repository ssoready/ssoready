package dsig

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ssoready/ssoready/internal/saml/c14n"
	"github.com/ssoready/ssoready/internal/saml/uxml"
)

var (
	ErrUnsigned       = fmt.Errorf("dsig: unsigned saml assertion")
	ErrNoRSAPublicKey = fmt.Errorf("dsig: cert does not contain *rsa.PublicKey")
	ErrBadDigest      = fmt.Errorf("dsig: digest mismatch in saml assertion")
)

type BadSignatureAlgorithmError struct {
	BadAlgorithm string
}

func (e BadSignatureAlgorithmError) Error() string {
	return fmt.Sprintf("dsig: bad signature algorithm: %s", e.BadAlgorithm)
}

type BadDigestAlgorithmError struct {
	BadAlgorithm string
}

func (e BadDigestAlgorithmError) Error() string {
	return fmt.Sprintf("dsig: bad digest algorithm: %s", e.BadAlgorithm)
}

type BadCertificateError struct {
	BadCertificate *x509.Certificate
}

func (e BadCertificateError) Error() string {
	return fmt.Sprintf("dsig: bad certificate on response")
}

func Verify(cert *x509.Certificate, data []byte) ([]byte, error) {
	unverifiedDoc, err := uxml.Parse(data)
	if err != nil {
		return nil, err
	}

	signatureValue, _ := onlyPathHoistNames(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:protocol", Local: "Response"},
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Signature"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "SignatureValue"},
	}, unverifiedDoc.Root)

	if signatureValue.Element == nil || signatureValue.Element.Children[0].Text == nil {
		return nil, ErrUnsigned
	}

	signatureBase64 := *signatureValue.Element.Children[0].Text

	signatureMethod, _ := onlyPathHoistNames(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:protocol", Local: "Response"},
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Signature"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "SignedInfo"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "SignatureMethod"},
	}, unverifiedDoc.Root)

	signatureMethodAlgorithm, _ := attrValueIgnoreNamespace(signatureMethod, "Algorithm")
	if signatureMethodAlgorithm != "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256" {
		return nil, BadSignatureAlgorithmError{signatureMethodAlgorithm}
	}

	digestMethod, _ := onlyPathHoistNames(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:protocol", Local: "Response"},
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Signature"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "SignedInfo"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Reference"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "DigestMethod"},
	}, unverifiedDoc.Root)

	digestMethodAlgorithm, _ := attrValueIgnoreNamespace(digestMethod, "Algorithm")
	if digestMethodAlgorithm != "http://www.w3.org/2001/04/xmlenc#sha256" {
		return nil, BadDigestAlgorithmError{digestMethodAlgorithm}
	}

	x509Certificate, _ := onlyPathHoistNames(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:protocol", Local: "Response"},
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Signature"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "KeyInfo"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "X509Data"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "X509Certificate"},
	}, unverifiedDoc.Root)

	resCertBase64 := *x509Certificate.Element.Children[0].Text
	resCertBase64 = strings.ReplaceAll(resCertBase64, " ", "")
	resCertBase64 = strings.ReplaceAll(resCertBase64, "\n", "")
	resCertRaw, err := base64.StdEncoding.DecodeString(resCertBase64)
	if err != nil {
		return nil, fmt.Errorf("parse saml response certificate: %w", err)
	}

	if !bytes.Equal(resCertRaw, cert.Raw) {
		badCert, err := x509.ParseCertificate(resCertRaw)
		if err != nil {
			return nil, fmt.Errorf("parse saml response certificate: %w", err)
		}

		return nil, BadCertificateError{BadCertificate: badCert}
	}

	digestData, err := responseDigestData(unverifiedDoc)
	if err != nil {
		return nil, err
	}

	digestHash := sha256.Sum256(digestData)
	digestHashBase64 := base64.StdEncoding.EncodeToString(digestHash[:])

	digestValue, _ := onlyPathHoistNames(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:protocol", Local: "Response"},
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Signature"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "SignedInfo"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Reference"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "DigestValue"},
	}, unverifiedDoc.Root)

	if *digestValue.Element.Children[0].Text != digestHashBase64 {
		return nil, ErrBadDigest
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, ErrNoRSAPublicKey
	}

	signatureData, err := responseSignatureData(data)
	if err != nil {
		return nil, err
	}

	signatureHash := sha256.Sum256(signatureData)

	//signatureBase64 := signatureBase64
	signatureBase64 = strings.ReplaceAll(signatureBase64, " ", "")
	signatureBase64 = strings.ReplaceAll(signatureBase64, "\n", "")
	expectedSignature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return nil, err
	}

	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, signatureHash[:], expectedSignature); err != nil {
		return nil, fmt.Errorf("verify signature: %w", err)
	}

	return digestData, nil
}

func responseDigestData(unverifiedDoc *uxml.Document) ([]byte, error) {
	assertion, _ := onlyPathHoistNames(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:protocol", Local: "Response"},
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
	}, unverifiedDoc.Root)

	nosig := exceptPath(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Signature"},
	}, assertion)

	transforms, _ := onlyPathHoistNames(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:protocol", Local: "Response"},
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Signature"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "SignedInfo"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Reference"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Transforms"},
	}, unverifiedDoc.Root)

	var inclusiveNamespaces []string
	for _, t := range transforms.Element.Children {
		algorithm, _ := attrValueIgnoreNamespace(t, "Algorithm")
		if algorithm != "http://www.w3.org/2001/10/xml-exc-c14n#" {
			continue
		}

		var inclusiveNamespacesElement uxml.Node
		for _, c := range t.Element.Children {
			if c.Element.Name.Local == "InclusiveNamespaces" {
				inclusiveNamespacesElement = c
			}
		}
		prefixList, _ := attrValueIgnoreNamespace(inclusiveNamespacesElement, "PrefixList")

		// ensure inclusiveNamespaces remains empty if PrefixList is empty
		// ("empty" here likely just means the assertion XML lacks InclusiveNamespaces at all)
		if prefixList != "" {
			inclusiveNamespaces = strings.Split(prefixList, " ")
		}
	}

	return c14n.Canonicalize(nosig, inclusiveNamespaces)
}

func responseSignatureData(data []byte) ([]byte, error) {
	doc, err := uxml.Parse(data)
	if err != nil {
		return nil, err
	}

	// todo remove ok?
	n, _ := onlyPathHoistNames(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:protocol", Local: "Response"},
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Signature"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "SignedInfo"},
	}, doc.Root)

	return c14n.Canonicalize(n, nil)
}
