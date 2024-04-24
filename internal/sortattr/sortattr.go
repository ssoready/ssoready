package sortattr

// todo this might be better named "sortattrname"

import (
	"github.com/ssoready/ssoready/internal/uxml"
)

// SortAttr can sort attributes in compliance with the c14n specification.
type SortAttr struct {
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

	spaceI, isSpaceI := s.Attrs[i].Name.Space()
	spaceJ, isSpaceJ := s.Attrs[j].Name.Space()

	// It follows that the very first node is the default namespace node. Let's
	// handle those first:
	if isSpaceI && !isSpaceJ {
		return true
	}
	if !isSpaceI && isSpaceJ {
		return false
	}

	// Namespace nodes go first. If one is a namespace node and the other isn't,
	// then it goes first.
	if isSpaceI && !isSpaceJ {
		return true
	}
	if !isSpaceI && isSpaceJ {
		return false
	}

	// Break ties between two namespace nodes by their local name.
	if isSpaceI && isSpaceJ {
		return spaceI < spaceJ
	}

	// Finally:
	//
	// "An element's attribute nodes are sorted lexicographically with namespace
	// URI as the primary key and local name as the secondary key (an empty
	// namespace URI is lexicographically least)."

	if s.Attrs[i].Name.URI != s.Attrs[j].Name.URI {
		return s.Attrs[i].Name.URI < s.Attrs[j].Name.URI
	}
	return s.Attrs[i].Name.Local < s.Attrs[j].Name.Local
}
