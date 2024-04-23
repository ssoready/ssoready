package sortattr

// todo this might be better named "sortattrname"

import (
	"strings"

	"github.com/ssoready/ssoready/internal/uxml"
	"github.com/ssoready/ssoready/internal/uxml/stack"
)

// SortAttr can sort attributes in compliance with the c14n specification.
type SortAttr struct {
	Stack *stack.Stack
	Attrs []uxml.Attr
}

// Len implements Sort.
func (s SortAttr) Len() int {
	return len(s.Attrs)
}

// Swap implements Sort.
func (s SortAttr) Swap(i, j int) {
	s.Attrs[i], s.Attrs[j] = s.Attrs[j], s.Attrs[i]
}

// Less implements Sort.
func (s SortAttr) Less(i, j int) bool {
	// Many comments in this function are copied from:
	//
	// https://www.w3.org/TR/2001/REC-xml-c14n-20010315#DocumentOrder

	// The spec states:
	//
	// "Namespace nodes have a lesser document order position than attribute
	// nodes."
	//
	// And:
	//
	// "An element's namespace nodes are sorted lexicographically by local name
	// (the default namespace node, if one exists, has no local name and is
	// therefore lexicographically least)."
	//
	// It follows that the very first node is the default namespace node. Let's
	// handle those first:
	if s.Attrs[i].Name == "xmlns" {
		return true
	}
	if s.Attrs[j].Name == "xmlns" {
		return false
	}

	qualI, localI := splitName(s.Attrs[i].Name)
	qualJ, localJ := splitName(s.Attrs[j].Name)

	// Namespace nodes go first. If one is a namespace node and the other isn't,
	// then it goes first.
	if qualI == "xmlns" && qualJ != "xmlns" {
		return true
	}
	if qualI != "xmlns" && qualJ == "xmlns" {
		return false
	}

	// Break ties between two namespace nodes by their local name.
	if qualI == "xmlns" && qualJ == "xmlns" {
		return localI < localJ
	}

	// Finally:
	//
	// "An element's attribute nodes are sorted lexicographically with namespace
	// URI as the primary key and local name as the secondary key (an empty
	// namespace URI is lexicographically least)."

	spaceI, _ := s.Stack.Get(qualI)
	spaceJ, _ := s.Stack.Get(qualJ)
	if spaceI != spaceJ {
		return spaceI < spaceJ
	}

	return localI < localJ
}

func splitName(s string) (string, string) {
	i := strings.IndexByte(s, ':')
	if i == -1 {
		return "", s
	}
	return s[:i], s[i+1:]
}
