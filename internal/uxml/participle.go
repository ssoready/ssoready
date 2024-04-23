package uxml

import (
	"fmt"
	"strconv"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var parser = participle.MustBuild[doc](
	participle.Lexer(lexer.MustStateful(lexer.Rules{
		"Root": {
			{"<?", `<\?`, lexer.Push("Declaration")},
			{"<", `<`, lexer.Push("Element")},
			{"BeginText", `[^<]`, lexer.Push("Text")},
		},
		"Declaration": {
			{"?>", `\?>`, lexer.Pop()},
			{"=", `=`, nil},
			{"S", `[ \t\r\n]+`, nil},
			{"Name", `[a-zA-Z][a-zA-Z0-9:]*`, nil},
			{"String", `"([^"\\]|\\.)*"`, nil},
		},
		"Element": {
			{">", `>`, lexer.Pop()},
			{"<", `<`, lexer.Push("Element")},
			{"=", `=`, nil},
			{"/", `/`, nil},
			{"S", `[ \t\r\n]+`, nil},
			{"Name", `[a-zA-Z][a-zA-Z0-9:]*`, nil},
			{"String", `"([^"\\]|\\.)*"`, nil},
		},
		"Text": {
			{"Text", `[^<]+`, nil},
			lexer.Return(),
		},
	})),
	participle.Elide("S"),
	participle.Union[node](elem{}, text{}),
	participle.Union[elemTail](elemTailEmpty{}, elemTailChildren{}),
)

type doc struct {
	Declaration struct{} `parser:"('<?' Name (Name '=' String)* '?>')?"`
	Nodes       []node   `parser:"@@*"`
}

type node interface {
	node()
}

type elem struct {
	Name  string   `parser:"'<' @Name"`
	Attrs []attr   `parser:"@@*"`
	Tail  elemTail `parser:"@@"`
}

func (elem) node() {}

type elemTail interface {
	elemTail()
}

type elemTailEmpty struct {
	Empty string `parser:"'/' '>'"`
}

func (elemTailEmpty) elemTail() {}

type elemTailChildren struct {
	Children []node `parser:"'>' @@* '<' '/' Name '>'"`
}

func (elemTailChildren) elemTail() {}

type attr struct {
	Name  string `parser:"@Name '='"`
	Value string `parser:"@String"`
}

type text struct {
	Start string `parser:"@BeginText"`
	Rest  string `parser:"@Text?"`
}

func (text) node() {}

func convertDocument(d doc) (*Document, error) {
	for _, n := range d.Nodes {
		node, err := convertNode(n)
		if err != nil {
			return nil, err
		}
		if node.Element == nil {
			continue
		}
		return &Document{Root: *node}, nil
	}
	return nil, fmt.Errorf("doc has no element nodes")
}

func convertNode(n node) (*Node, error) {
	switch n := n.(type) {
	case elem:
		elem, err := convertElement(n)
		if err != nil {
			return nil, err
		}
		return &Node{Element: elem}, nil
	case text:
		text, err := convertText(n)
		if err != nil {
			return nil, err
		}
		return &Node{Text: text}, nil
	default:
		panic("unreachable")
	}
}

func convertElement(e elem) (*Element, error) {
	elem := Element{Name: e.Name}
	for _, a := range e.Attrs {
		attr, err := convertAttr(a)
		if err != nil {
			return nil, err
		}
		elem.Attrs = append(elem.Attrs, *attr)
	}
	switch t := e.Tail.(type) {
	case elemTailEmpty:
		// no-op
	case elemTailChildren:
		for _, c := range t.Children {
			node, err := convertNode(c)
			if err != nil {
				return nil, err
			}
			elem.Children = append(elem.Children, *node)
		}
	}
	return &elem, nil
}

func convertAttr(a attr) (*Attr, error) {
	s, err := decodeEntities(a.Value[1 : len(a.Value)-1])
	if err != nil {
		return nil, err
	}
	return &Attr{Name: a.Name, Value: s}, nil
}

func convertText(t text) (*string, error) {
	s, err := decodeEntities(t.Start + t.Rest)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func decodeEntities(s string) (string, error) {
	var out []byte
	var i int
	for i < len(s) {
		// handle non-entity case
		if s[i] != '&' {
			out = append(out, s[i])
			i++
			continue
		}

		// scan to end of entity
		var j int
		for j = i; j < len(s); j++ {
			if s[j] == ';' {
				break
			}
		}

		if s[j] != ';' {
			return "", fmt.Errorf("unterminated entity: %q", s)
		}

		val, err := decodeSingleEntity(s[i+1 : j])
		if err != nil {
			return "", err
		}
		out = append(out, []byte(val)...)
		i = j + 1
	}
	return string(out), nil
}

func decodeSingleEntity(s string) (string, error) {
	switch s {
	case "lt":
		return "<", nil
	case "gt":
		return ">", nil
	case "amp":
		return "&", nil
	case "apos":
		return "'", nil
	case "quot":
		return "\"", nil
	}

	if len(s) < 2 || s[0] != '#' {
		return "", fmt.Errorf("invalid entity: %q", s)
	}

	parse := s[1:]
	base := 10
	if s[1] == 'x' {
		parse = s[2:]
		base = 16
	}

	n, err := strconv.ParseInt(parse, base, 0)
	if err != nil {
		return "", fmt.Errorf("invalid numeric entity: %w", err)
	}

	return string(rune(n)), nil
}
