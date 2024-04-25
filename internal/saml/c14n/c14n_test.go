package c14n_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ssoready/ssoready/internal/saml/c14n"
	"github.com/ssoready/ssoready/internal/saml/uxml"
	"github.com/stretchr/testify/assert"
)

func ExampleCanonicalize() {
	input := `<foo z="2" a="1"><bar /></foo>`
	doc, err := uxml.Parse([]byte(input))
	if err != nil {
		panic(err)
	}
	out, err := c14n.Canonicalize(doc.Root, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	// Output:
	// <foo a="1" z="2"><bar></bar></foo>
}

func TestCanonicalize(t *testing.T) {
	// Note:
	// tests/charmods is modified from the c14n test suite. CDATA is removed, because uxml intentionally does not
	// support CDATA. Single-quoted attributes are changed to double-quoted ones, for the same reason.

	entries, err := os.ReadDir("testdata")
	assert.NoError(t, err)

	for _, file := range entries {
		t.Run(file.Name(), func(t *testing.T) {
			in, err := os.ReadFile(fmt.Sprintf("testdata/%s/in.xml", file.Name()))
			assert.NoError(t, err)

			out, err := os.ReadFile(fmt.Sprintf("testdata/%s/out.xml", file.Name()))
			assert.NoError(t, err)

			doc, err := uxml.Parse(in)
			assert.NoError(t, err)

			actual, err := c14n.Canonicalize(doc.Root, nil)
			assert.NoError(t, err)
			assert.Equal(t, out, actual)
		})
	}
}
