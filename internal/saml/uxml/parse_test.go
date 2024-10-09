package uxml_test

import (
	"github.com/ssoready/ssoready/internal/saml/uxml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func textNode(s string) uxml.Node {
	return uxml.Node{Text: &s}
}

func TestParse(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  *uxml.Document
	}{
		{
			name: "single tag",
			in:   "<a></a>",
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a"}}}},
		},
		{
			name: "single tag, self-closing",
			in:   "<a />",
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a"}}}},
		},
		{
			name: "element with single attribute",
			in:   `<a href="https://example.com" />`,
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a"}, Attrs: []uxml.Attr{{Name: uxml.Name{Local: "href"}, Value: "https://example.com"}}}}},
		},
		{
			name: "element with multiple attributes",
			in:   `<a href="https://example.com" title="example" />`,
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a"}, Attrs: []uxml.Attr{{Name: uxml.Name{Local: "href"}, Value: "https://example.com"}, {Name: uxml.Name{Local: "title"}, Value: "example"}}}}},
		},
		{
			name: "element with text node child",
			in:   `<a>text</a>`,
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a"}, Children: []uxml.Node{textNode("text")}}}},
		},
		{
			name: "element with text node and attributes",
			in:   `<a href="https://example.com">text</a>`,
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a"}, Attrs: []uxml.Attr{{Name: uxml.Name{Local: "href"}, Value: "https://example.com"}}, Children: []uxml.Node{textNode("text")}}}},
		},
		{
			name: "element with default namespace declaration",
			in:   `<a xmlns="http://example.com" />`,
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a", URI: "http://example.com"}, Attrs: []uxml.Attr{{Name: uxml.Name{Local: "xmlns"}, Value: "http://example.com"}}}}},
		},
		{
			name: "element with qualified namespace declaration",
			in:   `<a xmlns:x="http://example.com" x:attr="value"/>`,
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a"}, Attrs: []uxml.Attr{{Name: uxml.Name{Qual: "xmlns", Local: "x"}, Value: "http://example.com"}, {Name: uxml.Name{Local: "attr", Qual: "x", URI: "http://example.com"}, Value: "value"}}}}},
		},
		{
			name: "element with qualified namespace declaration in child",
			in:   `<a xmlns:x="http://example.com"><x:child>value</x:child></a>`,
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a"}, Attrs: []uxml.Attr{{Name: uxml.Name{Qual: "xmlns", Local: "x"}, Value: "http://example.com"}}, Children: []uxml.Node{{Element: &uxml.Element{Name: uxml.Name{Local: "child", Qual: "x", URI: "http://example.com"}, Children: []uxml.Node{textNode("value")}}}}}}},
		},
		{
			name: "element using its own qualified namespace declaration",
			in:   `<x:a xmlns:x="http://example.com" x:attr="value"/>`,
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Qual: "x", Local: "a", URI: "http://example.com"}, Attrs: []uxml.Attr{{Name: uxml.Name{Qual: "xmlns", Local: "x"}, Value: "http://example.com"}, {Name: uxml.Name{Qual: "x", Local: "attr", URI: "http://example.com"}, Value: "value"}}}}},
		},
		{
			name: "nested elements",
			in:   `<a><b><c>text</c></b></a>`,
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a"}, Children: []uxml.Node{{Element: &uxml.Element{Name: uxml.Name{Local: "b"}, Children: []uxml.Node{{Element: &uxml.Element{Name: uxml.Name{Local: "c"}, Children: []uxml.Node{textNode("text")}}}}}}}}}},
		},
		{
			name: "element with nested qualified namespace declaration, inner declaration wins",
			in:   `<a xmlns:x="http://example.com"><b xmlns:x="http://different.com"><x:c>value</x:c></b></a>`,
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a"}, Attrs: []uxml.Attr{{Name: uxml.Name{Qual: "xmlns", Local: "x"}, Value: "http://example.com"}}, Children: []uxml.Node{{Element: &uxml.Element{Name: uxml.Name{Local: "b"}, Attrs: []uxml.Attr{{Name: uxml.Name{Qual: "xmlns", Local: "x"}, Value: "http://different.com"}}, Children: []uxml.Node{{Element: &uxml.Element{Name: uxml.Name{Local: "c", Qual: "x", URI: "http://different.com"}, Children: []uxml.Node{textNode("value")}}}}}}}}}},
		},
		{
			name: "element with two children using x namespace, b redeclares, c does not",
			in:   `<a xmlns:x="http://example.com"><x:b xmlns:x="http://different.com"></x:b><x:c></x:c></a>`,
			out: &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a"}, Attrs: []uxml.Attr{{Name: uxml.Name{Qual: "xmlns", Local: "x"}, Value: "http://example.com"}}, Children: []uxml.Node{
				{Element: &uxml.Element{Name: uxml.Name{Local: "b", Qual: "x", URI: "http://different.com"}, Attrs: []uxml.Attr{{Name: uxml.Name{Qual: "xmlns", Local: "x"}, Value: "http://different.com"}}}},
				{Element: &uxml.Element{Name: uxml.Name{Local: "c", Qual: "x", URI: "http://example.com"}}},
			}}}},
		},
		{
			name: "element with entity replacements",
			in:   `<a attr="&lt; &gt; &amp; &apos; &quot;">text &lt; &gt; &amp; &apos; &quot;</a>`,
			out: &uxml.Document{
				Root: uxml.Node{
					Element: &uxml.Element{
						Name: uxml.Name{Local: "a"},
						Attrs: []uxml.Attr{
							{Name: uxml.Name{Local: "attr"}, Value: `< > & ' "`},
						},
						Children: []uxml.Node{
							textNode(`text < > & ' "`),
						},
					},
				},
			},
		},
		{
			name: "element with numeric entity replacements",
			in:   `<a attr="&#60; &#62; &#38; &#39; &#34;">text &#60; &#62; &#38; &#39; &#34;</a>`,
			out: &uxml.Document{
				Root: uxml.Node{
					Element: &uxml.Element{
						Name: uxml.Name{Local: "a"},
						Attrs: []uxml.Attr{
							{Name: uxml.Name{Local: "attr"}, Value: `< > & ' "`},
						},
						Children: []uxml.Node{
							textNode(`text < > & ' "`),
						},
					},
				},
			},
		},
		{
			name: "declaration",
			in:   `<?xml version="1.0" encoding="UTF-8"?><a />`,
			out:  &uxml.Document{Root: uxml.Node{Element: &uxml.Element{Name: uxml.Name{Local: "a"}}}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			out, err := uxml.Parse([]byte(tt.in))
			assert.NoError(t, err)
			assert.Equal(t, tt.out, out)
		})
	}
}
