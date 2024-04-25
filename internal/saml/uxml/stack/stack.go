package stack

// Stack is a stack of XML namespace declarations.
type Stack []map[string]string

// Push pushes a set of names and their corresponding URIs to the top of the
// stack.
func (s *Stack) Push(names map[string]string) {
	*s = append(*s, names)
}

// Get fetches the URI for a name (or the empth string if not found) and whether
// the name was found at all. Definitions closer to the top of the stack take
// predence over values further from the top.
func (s *Stack) Get(name string) (string, bool) {
	for i := len(*s) - 1; i >= 0; i-- {
		if uri, ok := (*s)[i][name]; ok {
			return uri, true
		}
	}

	return "", false
}

// Pop pops the top of the name stack.
func (s *Stack) Pop() {
	(*s) = (*s)[:len(*s)-1]
}

// Len returns depth of the stack.
func (s *Stack) Len() int {
	return len(*s)
}

// GetAll returns all names in the stack, and their current values. Definitions
// closer to the top of the stack take predence over values further from the
// top.
func (s *Stack) GetAll() map[string]string {
	out := map[string]string{}
	for _, names := range *s {
		for name, uri := range names {
			out[name] = uri
		}
	}

	return out
}
