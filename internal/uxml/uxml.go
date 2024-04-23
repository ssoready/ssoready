// Package uxml implements a subset of XML for SAML.
package uxml

type Document struct {
	Root Node
}

type Node struct {
	Element *Element
	Text    *string
}

type Element struct {
	Name     string
	Attrs    []Attr
	Children []Node
}

type Attr struct {
	Name  string
	Value string
}

func Parse(s string) (*Document, error) {
	parseDoc, err := parser.ParseString("", s)
	if err != nil {
		return nil, err
	}
	doc, err := convertDocument(*parseDoc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
