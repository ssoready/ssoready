package secretload

import "unicode"

// copied from https://github.com/ssoready/conf/blob/master/envify.go

// envify returns the environment variable name for the flag named s.
//
// Effectively, envify calculates the SCREAMING_SNAKE_CASE rendition of s.
func envify(s string) string {
	in := []rune(s)
	var out []rune

	// Scan for words, starting from the end. A reverse search is convenient
	// here because it's easier to tell how to separate "JSONData" into
	// "JSON_DATA" when working backwards.
	//
	// When searching forward, you have to look ahead to see if an upper is
	// followed by a lower. When searching backward, you just have to allow a
	// lower word to be terminated by an upper.
	//
	// This idea is copied from segmentio/conf's snakecase.go
	i := len(in) - 1
	for i >= 0 {
		if unicode.IsLower(in[i]) {
			// Scan for a sequence of lowers.
			for i >= 0 && unicode.IsLower(in[i]) {
				out = append(out, unicode.ToUpper(in[i]))
				i--
			}

			// Insert the next letter, if it's an upper.
			if i >= 0 && unicode.IsUpper(in[i]) {
				out = append(out, in[i])
				i--
			}

			// Separate this word from the next, unless the next char is already
			// a separator.
			if i >= 0 && !unicode.IsPunct(in[i]) {
				out = append(out, '_')
			}
		} else if unicode.IsUpper(in[i]) {
			// Scan for a sequence of uppers.
			for i >= 0 && unicode.IsUpper(in[i]) {
				out = append(out, in[i])
				i--
			}

			// Separate this word from the next, unless the next char is already
			// a separator.
			if i >= 0 && !unicode.IsPunct(in[i]) {
				out = append(out, '_')
			}
		} else if unicode.IsPunct(in[i]) {
			// Replace this separator with an ASCII underscore.
			out = append(out, '_')
			i--
		} else {
			// Some non-letter, non-separator character. Just leave it as-is.
			out = append(out, in[i])
			i--
		}
	}

	// Reverse out, since we did a backwards search.
	for i, j := 0, len(out)-1; i < j; {
		out[i], out[j] = out[j], out[i]
		i++
		j--
	}

	return string(out)
}
