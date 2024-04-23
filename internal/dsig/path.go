package dsig

import (
	"strings"

	"github.com/ssoready/ssoready/internal/uxml"
	"github.com/ssoready/ssoready/internal/uxml/stack"
)

type path []segment

type segment struct {
	Namespace   string
	ElementName string
}

func walk(n uxml.Node, f func(path, stack.Stack, uxml.Node)) {
	walkInternal(nil, nil, n, f)
}

func walkInternal(p path, s stack.Stack, n uxml.Node, f func(path, stack.Stack, uxml.Node)) {
	if n.Text != nil {
		f(p, s, n)
		return
	}

	names := map[string]string{}
	for _, a := range n.Element.Attrs {
		ns, ok := strings.CutPrefix(a.Name, "xmlns:")
		if !ok {
			continue
		}
		names[ns] = a.Value
	}

	p = append(p)
	s.Push(names)
}
