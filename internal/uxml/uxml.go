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
	Name     Name
	Attrs    []Attr
	Children []Node
}

type Attr struct {
	Name  Name
	Value string
}

type Name struct {
	URI   string
	Qual  string
	Local string
}

func (n Name) Space() (string, bool) {
	if n.Qual == "" && n.Local == "xmlns" {
		return "", true
	}
	if n.Qual == "xmlns" {
		return n.Local, true
	}
	return "", false
}

func Parse(b []byte) (*Document, error) {
	parseDoc, err := parser.ParseBytes("", b)
	if err != nil {
		return nil, err
	}
	doc, err := convertDocument(*parseDoc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
