package dsig

import (
	"fmt"
	"os"
	"testing"

	"github.com/ssoready/ssoready/internal/c14n"
	"github.com/ssoready/ssoready/internal/uxml"
	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	in, err := os.ReadFile(fmt.Sprintf("../testdata/assertion-okta.xml"))
	assert.NoError(t, err)

	doc, err := uxml.Parse(string(in))
	assert.NoError(t, err)

	assertion, _ := onlyPath(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:protocol", Local: "Response"},
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
	}, doc.Root)
	fmt.Println("assertion", assertion.Element)
	b, err := c14n.Canonicalize(assertion, nil)
	//fmt.Println("c14n", string(b), err)

	nosig := exceptPath(path{
		{URI: "urn:oasis:names:tc:SAML:2.0:assertion", Local: "Assertion"},
		{URI: "http://www.w3.org/2000/09/xmldsig#", Local: "Signature"},
	}, assertion)

	b, err = c14n.Canonicalize(nosig, nil)
	fmt.Println("c14n", string(b), err)

	panic("a")
}
