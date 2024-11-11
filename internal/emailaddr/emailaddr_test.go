package emailaddr_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ssoready/ssoready/internal/emailaddr"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		in, out string
	}{
		{
			in:  "jdoe@example.com",
			out: "example.com",
		},
		{
			in:  "john.doe@example.com",
			out: "example.com",
		},
		{
			in:  "jdoe+foo@example.com",
			out: "example.com",
		},
		{
			in:  "jdoe@EXAMPLE.com",
			out: "example.com",
		},
		{
			in:  "john-doe@example.com",
			out: "example.com",
		},
		{
			in:  "john-doe.foo@example.com",
			out: "example.com",
		},
		/**
		 * In order to test for user group invites in Microsoft Entra, 
		 * we need to account for the addition of the #EXT# tag in invite emails.
		 * - https://github.com/ssoready/ssoready/issues/185
		 */
		{
			in: "john-doe#EXT#@example.com",
			out: "example.com",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.in, func(t *testing.T) {
			out, err := emailaddr.Parse(tt.in)
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if d := cmp.Diff(out, tt.out); d != "" {
				t.Fatalf("parse mismatch: (+want -got):\n%s", d)
			}
		})
	}
}
