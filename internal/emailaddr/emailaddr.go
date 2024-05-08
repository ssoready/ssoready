package emailaddr

import (
	"fmt"
	"regexp"
)

var pat = regexp.MustCompile(`^[a-zA-Z0-9_.]+@([a-zA-Z0-9_.]+)$`)

func Parse(s string) (string, error) {
	match := pat.FindStringSubmatch(s)
	if len(match) == 0 {
		return "", fmt.Errorf("invalid email address: %q", s)
	}

	return match[1], nil
}
