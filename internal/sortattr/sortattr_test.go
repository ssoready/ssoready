package sortattr_test

import (
	"sort"
	"strconv"
	"testing"

	"github.com/ssoready/ssoready/internal/sortattr"
	"github.com/ssoready/ssoready/internal/uxml"
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
					Name:  uxml.Name{Local: "xmlns"},
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  uxml.Name{Qual: "foo", Local: "bar"},
					Value: "baz",
				},
			},
			Out: []uxml.Attr{
				uxml.Attr{
					Name:  uxml.Name{Local: "xmlns"},
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  uxml.Name{Qual: "foo", Local: "bar"},
					Value: "baz",
				},
			},
		},
		testCase{
			In: []uxml.Attr{
				uxml.Attr{
					Name:  uxml.Name{Qual: "foo", Local: "bar"},
					Value: "baz",
				},
				uxml.Attr{
					Name:  uxml.Name{Local: "xmlns"},
					Value: "https://example.com",
				},
			},
			Out: []uxml.Attr{
				uxml.Attr{
					Name:  uxml.Name{Local: "xmlns"},
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  uxml.Name{Qual: "foo", Local: "bar"},
					Value: "baz",
				},
			},
		},
		testCase{
			In: []uxml.Attr{
				uxml.Attr{
					Name:  uxml.Name{Qual: "xmlns", Local: "foo"},
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  uxml.Name{Qual: "foo", Local: "bar"},
					Value: "baz",
				},
			},
			Out: []uxml.Attr{
				uxml.Attr{
					Name:  uxml.Name{Qual: "xmlns", Local: "foo"},
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  uxml.Name{Qual: "foo", Local: "bar"},
					Value: "baz",
				},
			},
		},
		testCase{
			In: []uxml.Attr{
				uxml.Attr{
					Name:  uxml.Name{Qual: "foo", Local: "bar"},
					Value: "baz",
				},
				uxml.Attr{
					Name:  uxml.Name{Qual: "xmlns", Local: "foo"},
					Value: "https://example.com",
				},
			},
			Out: []uxml.Attr{
				uxml.Attr{
					Name:  uxml.Name{Qual: "xmlns", Local: "foo"},
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  uxml.Name{Qual: "foo", Local: "bar"},
					Value: "baz",
				},
			},
		},
		testCase{
			In: []uxml.Attr{
				uxml.Attr{
					Name:  uxml.Name{Qual: "xmlns", Local: "foo"},
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  uxml.Name{Qual: "xmlns", Local: "bar"},
					Value: "https://example.com",
				},
			},
			Out: []uxml.Attr{
				uxml.Attr{
					Name:  uxml.Name{Qual: "xmlns", Local: "bar"},
					Value: "https://example.com",
				},
				uxml.Attr{
					Name:  uxml.Name{Qual: "xmlns", Local: "foo"},
					Value: "https://example.com",
				},
			},
		},
		testCase{
			In: []uxml.Attr{
				uxml.Attr{
					Name:  uxml.Name{Qual: "a", Local: "attr", URI: "http://www.w3.org"},
					Value: "out",
				},
				uxml.Attr{
					Name:  uxml.Name{Qual: "b", Local: "attr", URI: "http://www.ietf.org"},
					Value: "sorted",
				},
				uxml.Attr{
					Name:  uxml.Name{Local: "attr2", URI: "http://www.example.com"},
					Value: "all",
				},
				uxml.Attr{
					Name:  uxml.Name{Local: "attr", URI: "http://www.example.com"},
					Value: "I'm",
				},
			},
			Out: []uxml.Attr{
				uxml.Attr{
					Name:  uxml.Name{Local: "attr", URI: "http://www.example.com"},
					Value: "I'm",
				},
				uxml.Attr{
					Name:  uxml.Name{Local: "attr2", URI: "http://www.example.com"},
					Value: "all",
				},
				uxml.Attr{
					Name:  uxml.Name{Qual: "b", Local: "attr", URI: "http://www.ietf.org"},
					Value: "sorted",
				},
				uxml.Attr{
					Name:  uxml.Name{Qual: "a", Local: "attr", URI: "http://www.w3.org"},
					Value: "out",
				},
			},
		},
	}

	for i, tt := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			attrs := sortattr.SortAttr{Attrs: tt.In}
			sort.Sort(attrs)
			assert.Equal(t, tt.Out, attrs.Attrs)
		})
	}
}

// <e5 a:attr="out" b:attr="sorted" attr2="all" attr="I'm"
// xmlns:b="http://www.ietf.org"
// xmlns:a="http://www.w3.org"
// xmlns="http://example.org"/>
