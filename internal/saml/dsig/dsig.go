package dsig

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/ssoready/ssoready/internal/saml/c14n"
	"github.com/ssoready/ssoready/internal/saml/samltypes"
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

func Verify(cert *x509.Certificate, data []byte) error {
	var res samltypes.Response
	if err := xml.Unmarshal(data, &res); err != nil {
		return err
	}

	fmt.Printf("%+v\n", res)

	if res.Assertion.Signature.SignatureValue == "" {
		return ErrUnsigned
	}

	if res.Assertion.Signature.SignedInfo.SignatureMethod.Algorithm != "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256" {
		return BadSignatureAlgorithmError{res.Assertion.Signature.SignedInfo.SignatureMethod.Algorithm}
	}

	if res.Assertion.Signature.SignedInfo.Reference.DigestMethod.Algorithm != "http://www.w3.org/2001/04/xmlenc#sha256" {
		return BadDigestAlgorithmError{res.Assertion.Signature.SignedInfo.Reference.DigestMethod.Algorithm}
	}

	resCertBase64 := res.Assertion.Signature.KeyInfo.X509Data.X509Certificate
	resCertBase64 = strings.ReplaceAll(resCertBase64, " ", "")
	resCertBase64 = strings.ReplaceAll(resCertBase64, "\n", "")
	resCertRaw, err := base64.StdEncoding.DecodeString(resCertBase64)
	if err != nil {
		return fmt.Errorf("parse saml response certificate: %w", err)
	}

	if !bytes.Equal(resCertRaw, cert.Raw) {
		badCert, err := x509.ParseCertificate(resCertRaw)
		if err != nil {
			return fmt.Errorf("parse saml response certificate: %w", err)
		}

		return BadCertificateError{BadCertificate: badCert}
	}

	digestData, err := responseDigestData(res, data)
	if err != nil {
		return err
	}

	digestHash := sha256.Sum256(digestData)
	digestHashBase64 := base64.StdEncoding.EncodeToString(digestHash[:])

	if res.Assertion.Signature.SignedInfo.Reference.DigestValue != digestHashBase64 {
		return ErrBadDigest
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return ErrNoRSAPublicKey
	}

	signatureData, err := responseSignatureData(data)
	if err != nil {
		return err
	}

	signatureHash := sha256.Sum256(signatureData)

	signatureBase64 := res.Assertion.Signature.SignatureValue
	signatureBase64 = strings.ReplaceAll(signatureBase64, " ", "")
	signatureBase64 = strings.ReplaceAll(signatureBase64, "\n", "")
	expectedSignature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return err
	}

	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, signatureHash[:], expectedSignature); err != nil {
		return fmt.Errorf("verify signature: %w", err)
	}

	return nil
}

func responseDigestData(res samltypes.Response, data []byte) ([]byte, error) {
	doc, err := uxml.Parse(data)
	if err != nil {
		return nil, err
	}

	// todo hoist here, and get rid of onlyPath?
	assertion, _ := onlyPathHoistNames(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:protocol", Local: "Response"},
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
	}, doc.Root)

	nosig := exceptPath(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Signature"},
	}, assertion)

	var inclusiveNamespaces []string
	for _, t := range res.Assertion.Signature.SignedInfo.Reference.Transforms.Transform {
		// ensure inclusiveNamespaces remains empty if PrefixList is empty
		// ("empty" here likely just means the assertion XML lacks InclusiveNamespaces at all)
		if t.Algorithm == "http://www.w3.org/2001/10/xml-exc-c14n#" && t.InclusiveNamespaces.PrefixList != "" {
			inclusiveNamespaces = strings.Split(t.InclusiveNamespaces.PrefixList, " ")
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
