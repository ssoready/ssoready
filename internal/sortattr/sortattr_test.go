package sortattr_test

import (
	"sort"
	"strconv"
	"testing"

	"github.com/ssoready/ssoready/internal/sortattr"
	"github.com/ssoready/ssoready/internal/uxml"
	"github.com/ssoready/ssoready/internal/uxml/stack"
	"github.com/stretchr/testify/assert"
)

func TestSortAttr(t *testing.T) {
	type testCase struct {
		In  []uxml.Attr
		Out []uxml.Attr
	}

	testCases := []testCase{
		testCase{
			In: []uxml.Attr{
				uxml.Attr{
					Name:  "xmlns",
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  "foo:bar",
					Value: "baz",
				},
			},
			Out: []uxml.Attr{
				uxml.Attr{
					Name:  "xmlns",
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  "foo:bar",
					Value: "baz",
				},
			},
		},
		testCase{
			In: []uxml.Attr{
				uxml.Attr{
					Name:  "foo:bar",
					Value: "baz",
				},
				uxml.Attr{
					Name:  "xmlns",
					Value: "https://example.com",
				},
			},
			Out: []uxml.Attr{
				uxml.Attr{
					Name:  "xmlns",
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  "foo:bar",
					Value: "baz",
				},
			},
		},
		testCase{
			In: []uxml.Attr{
				uxml.Attr{
					Name:  "xmlns:foo",
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  "foo:bar",
					Value: "baz",
				},
			},
			Out: []uxml.Attr{
				uxml.Attr{
					Name:  "xmlns:foo",
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  "foo:bar",
					Value: "baz",
				},
			},
		},
		testCase{
			In: []uxml.Attr{
				uxml.Attr{
					Name:  "foo:bar",
					Value: "baz",
				},
				uxml.Attr{
					Name:  "xmlns:foo",
					Value: "https://example.com",
				},
			},
			Out: []uxml.Attr{
				uxml.Attr{
					Name:  "xmlns:foo",
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  "foo:bar",
					Value: "baz",
				},
			},
		},
		testCase{
			In: []uxml.Attr{
				uxml.Attr{
					Name:  "xmlns:foo",
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  "xmlns:bar",
					Value: "https://example.com",
				},
			},
			Out: []uxml.Attr{
				uxml.Attr{
					Name:  "xmlns:bar",
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  "xmlns:foo",
					Value: "https://example.com",
				},
			},
		},
		testCase{
			In: []uxml.Attr{
				uxml.Attr{
					Name:  "a:attr",
					Value: "out",
				},
				uxml.Attr{
					Name:  "b:attr",
					Value: "sorted",
				},
				uxml.Attr{
					Name:  "attr2",
					Value: "all",
				},
				uxml.Attr{
					Name:  "attr",
					Value: "I'm",
				},
			},
			Out: []uxml.Attr{
				uxml.Attr{
					Name:  "attr",
					Value: "I'm",
				},
				uxml.Attr{
					Name:  "attr2",
					Value: "all",
				},
				uxml.Attr{
					Name:  "b:attr",
					Value: "sorted",
				},
				uxml.Attr{
					Name:  "a:attr",
					Value: "out",
				},
			},
		},
	}

	var s stack.Stack
	s.Push(map[string]string{
		"":  "http://example.com",
		"a": "http://www.w3.org",
		"b": "http://www.ietf.org",
	})

	for i, tt := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			attrs := sortattr.SortAttr{Attrs: tt.In, Stack: &s}
			sort.Sort(attrs)
			assert.Equal(t, tt.Out, attrs.Attrs)
		})
	}
}

// <e5 a:attr="out" b:attr="sorted" attr2="all" attr="I'm"
// xmlns:b="http://www.ietf.org"
// xmlns:a="http://www.w3.org"
// xmlns="http://example.org"/>
