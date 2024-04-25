package dsig

import (
	"github.com/ssoready/ssoready/internal/saml/uxml"
	"github.com/ssoready/ssoready/internal/saml/uxml/stack"
)

type path []segment

type segment struct {
	URI   string
	Local string
}

func onlyPath(p path, n uxml.Node) (uxml.Node, bool) {
	if len(p) == 0 {
		panic("empty path")
	}
	if n.Element == nil {
		return uxml.Node{}, false
	}
	if n.Element.Name.URI != p[0].URI || n.Element.Name.Local != p[0].Local {
		return uxml.Node{}, false
	}
	if len(p) == 1 {
		return n, true
	}
	for _, c := range n.Element.Children {
		if find, ok := onlyPath(p[1:], c); ok {
			return find, true
		}
	}
	return uxml.Node{}, false
}

func onlyPathHoistNames(p path, n uxml.Node) (uxml.Node, bool) {
	return onlyPathHoistNamesInternal(p, stack.Stack{}, n)
}

func onlyPathHoistNamesInternal(p path, s stack.Stack, n uxml.Node) (uxml.Node, bool) {
	if len(p) == 0 {
		panic("empty path")
	}
	if n.Element == nil {
		return uxml.Node{}, false
	}
	if n.Element.Name.URI != p[0].URI || n.Element.Name.Local != p[0].Local {
		return uxml.Node{}, false
	}

	names := map[string]string{}
	for _, a := range n.Element.Attrs {
		if space, ok := a.Name.Space(); ok {
			names[space] = a.Value
		}
	}
	s.Push(names)
	defer s.Pop()

	if len(p) == 1 {
		var attrs []uxml.Attr
		for _, a := range n.Element.Attrs {
			attrs = append(attrs, a)
		}
		for k, v := range s.GetAll() {
			if k == "" {
				attrs = append(attrs, uxml.Attr{
					Name:  uxml.Name{Local: "xmlns"},
					Value: v,
				})
			} else {
				attrs = append(attrs, uxml.Attr{
					Name:  uxml.Name{Qual: "xmlns", Local: k},
					Value: v,
				})
			}
		}
		return uxml.Node{Element: &uxml.Element{
			Name:     n.Element.Name,
			Attrs:    attrs,
			Children: n.Element.Children,
		}}, true
	}

	for _, c := range n.Element.Children {
		if find, ok := onlyPathHoistNamesInternal(p[1:], s, c); ok {
			return find, true
		}
	}

	return uxml.Node{}, false
}

func exceptPath(p path, n uxml.Node) uxml.Node {
	// todo handle n being text
	var cur path
	cur = append(cur, segment{URI: n.Element.Name.URI, Local: n.Element.Name.Local})
	return exceptPathInternal(p, cur, n)
}

func exceptPathInternal(p, cur path, n uxml.Node) uxml.Node {
	if n.Element == nil {
		return n
	}

	var children []uxml.Node
	for _, c := range n.Element.Children {
		if c.Element != nil && pathsEqual(p, append(cur, segment{URI: c.Element.Name.URI, Local: c.Element.Name.Local})) {
			continue
		}
		children = append(children, exceptPathInternal(p, cur, c))
	}

	return uxml.Node{
		Element: &uxml.Element{
			Name:     n.Element.Name,
			Attrs:    n.Element.Attrs,
			Children: children,
		},
	}
}

func pathsEqual(a, b path) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i].URI != b[i].URI || a[i].Local != b[i].Local {
			return false
		}
	}
	return true
}
