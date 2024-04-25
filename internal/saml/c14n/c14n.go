package c14n

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/ssoready/ssoready/internal/saml/sortattr"
	"github.com/ssoready/ssoready/internal/saml/uxml"
	"github.com/ssoready/ssoready/internal/saml/uxml/stack"
)

func Canonicalize(n uxml.Node, inclusiveNamespaces []string) ([]byte, error) {
	var buf bytes.Buffer
	var knownNames, renderedNames stack.Stack
	canonicalize(&buf, knownNames, renderedNames, n, inclusiveNamespaces)
	return buf.Bytes(), nil
}

func canonicalize(buf *bytes.Buffer, knownNames, renderedNames stack.Stack, n uxml.Node, inclusiveNamespaces []string) {
	if n.Text != nil {
		t := []byte(*n.Text)
		t = bytes.ReplaceAll(t, amp, escAmp)
		t = bytes.ReplaceAll(t, lt, escLt)
		t = bytes.ReplaceAll(t, gt, escGt)
		t = bytes.ReplaceAll(t, cr, escCr)

		_, _ = fmt.Fprintf(buf, "%s", t)
		return
	}

	// Note the previous value of the default namespace. This needs to be
	// special-cased because the c14n spec special-cases the case of xmlns="".
	previousDefaultNamespace, _ := knownNames.Get("")

	names := map[string]string{} // names declared by this element
	for _, a := range n.Element.Attrs {
		if space, ok := a.Name.Space(); ok {
			names[space] = a.Value
		}
	}

	knownNames.Push(names)

	visiblyUsedNames := map[string]struct{}{} // names visibly used by this element

	// inclusiveNamespaces are always visibly used
	for _, s := range inclusiveNamespaces {
		visiblyUsedNames[s] = struct{}{}
	}

	visiblyUsedNames[n.Element.Name.Qual] = struct{}{}
	for _, a := range n.Element.Attrs {
		if _, ok := a.Name.Space(); !ok {
			visiblyUsedNames[a.Name.Qual] = struct{}{}
		}
	}

	namesToRender := map[string]struct{}{}
	for name, uri := range knownNames.GetAll() {
		var shouldRender bool
		if name == "" && uri == "" {
			_, visiblyUsed := visiblyUsedNames[""]
			declaredValue, declared := names[""]
			_, rendered := renderedNames.Get("")

			shouldRender = visiblyUsed && (!declared || declaredValue != previousDefaultNamespace) && rendered
		} else {
			_, visiblyUsed := visiblyUsedNames[name]
			renderedValue, rendered := renderedNames.Get(name)

			shouldRender = visiblyUsed && (!rendered || renderedValue != uri)
		}

		if shouldRender {
			namesToRender[name] = struct{}{}
		}
	}

	// attrsToRender is the set of attributes we'll render. The order doesn't
	// matter yet, we'll sort them later.
	var attrsToRender []uxml.Attr
	renderedNameValues := map[string]string{}

	// first, add all non-namespace attrs to attrsToRender
	for _, a := range n.Element.Attrs {
		if _, ok := a.Name.Space(); !ok {
			attrsToRender = append(attrsToRender, a)
		}
	}

	// next, go through namesToRender and add them to renderedNameValues and attrsToRender
	for name := range namesToRender {
		uri, _ := knownNames.Get(name)
		renderedNameValues[name] = uri

		if name == "" {
			attrsToRender = append(attrsToRender, uxml.Attr{
				Name: uxml.Name{
					Local: "xmlns",
				},
				Value: uri,
			})
		} else {
			attrsToRender = append(attrsToRender, uxml.Attr{
				Name: uxml.Name{
					Qual:  "xmlns",
					Local: name,
				},
				Value: uri,
			})
		}
	}

	renderedNames.Push(renderedNameValues)

	sortAttr := sortattr.SortAttr{Attrs: attrsToRender}
	sort.Sort(sortAttr)

	if n.Element.Name.Qual == "" {
		_, _ = fmt.Fprintf(buf, "<%s", n.Element.Name.Local)
	} else {
		_, _ = fmt.Fprintf(buf, "<%s:%s", n.Element.Name.Qual, n.Element.Name.Local)
	}

	for _, a := range sortAttr.Attrs {
		if a.Name.Qual == "" {
			_, _ = fmt.Fprintf(buf, " %s=\"", a.Name.Local)
		} else {
			_, _ = fmt.Fprintf(buf, " %s:%s=\"", a.Name.Qual, a.Name.Local)
		}

		val := []byte(a.Value)
		val = bytes.ReplaceAll(val, amp, escAmp)
		val = bytes.ReplaceAll(val, lt, escLt)
		val = bytes.ReplaceAll(val, quot, escQuot)
		val = bytes.ReplaceAll(val, tab, escTab)
		val = bytes.ReplaceAll(val, nl, escNl)
		val = bytes.ReplaceAll(val, cr, escCr)
		_, _ = fmt.Fprintf(buf, "%s\"", val)
	}

	_, _ = fmt.Fprint(buf, ">")

	for _, c := range n.Element.Children {
		canonicalize(buf, knownNames, renderedNames, c, inclusiveNamespaces)
	}

	if n.Element.Name.Qual == "" {
		_, _ = fmt.Fprintf(buf, "</%s>", n.Element.Name.Local)
	} else {
		_, _ = fmt.Fprintf(buf, "</%s:%s>", n.Element.Name.Qual, n.Element.Name.Local)
	}

	knownNames.Pop()
	renderedNames.Pop()
}

var (
	amp     = []byte("&")
	escAmp  = []byte("&amp;")
	lt      = []byte("<")
	escLt   = []byte("&lt;")
	gt      = []byte(">")
	escGt   = []byte("&gt;")
	cr      = []byte("\r")
	escCr   = []byte("&#xD;")
	quot    = []byte("\"")
	escQuot = []byte("&quot;")
	tab     = []byte("\t")
	escTab  = []byte("&#x9;")
	nl      = []byte("\n")
	escNl   = []byte("&#xA;")
)
