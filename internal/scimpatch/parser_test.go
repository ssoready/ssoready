package scimpatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePath(t *testing.T) {
	testCases := []struct {
		in  string
		out []pathSegment
	}{
		{
			in:  "",
			out: nil,
		},
		{
			in: "foo",
			out: []pathSegment{
				{name: "foo"},
			},
		},
		{
			in: "foo.bar",
			out: []pathSegment{
				{name: "foo"},
				{name: "bar"},
			},
		},
		{
			in: `a[b eq "c"]`,
			out: []pathSegment{
				{name: "a", filter: &filterExpr{attr: "b", op: "eq", value: "c"}},
			},
		},
		{
			in: `a[b pr]`,
			out: []pathSegment{
				{name: "a", filter: &filterExpr{attr: "b", op: "pr"}},
			},
		},
		{
			in: `a[b eq "c"].d`,
			out: []pathSegment{
				{name: "a", filter: &filterExpr{attr: "b", op: "eq", value: "c"}},
				{name: "d"},
			},
		},
		{
			in: `a[b pr].d`,
			out: []pathSegment{
				{name: "a", filter: &filterExpr{attr: "b", op: "pr"}},
				{name: "d"},
			},
		},

		{
			in: "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
			out: []pathSegment{
				{name: "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"},
			},
		},
		{
			in: "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department[type eq \"Engineering\"]",
			out: []pathSegment{
				{name: "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"},
				{name: "department", filter: &filterExpr{attr: "type", op: "eq", value: "Engineering"}},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.out, splitPath(tt.in))
		})
	}
}
