// Package trailer parses and formats RFC 822-style commit message trailers.
//
// A commit message is treated as a body (free text) optionally followed by
// a single trailer block as its last paragraph. The trailer block is a run
// of lines where every non-empty line either matches `Key: value` or is a
// folded continuation (starts with whitespace). This is a deliberately
// strict subset of git's tolerance rule; for everything that needs to match
// git exactly (composition of new trailers) callers should delegate to
// `git interpret-trailers` rather than re-emit through this package.
package trailer

import (
	"regexp"
	"strings"
)

// Trailer is a single Key/Value entry from a commit message trailer block.
type Trailer struct {
	Key   string
	Value string
}

var trailerLine = regexp.MustCompile(`^[A-Za-z0-9-]+:( |$)`)

// Parse splits a commit message into body and trailers. If no trailer block
// is detected, the entire (right-trimmed) message is returned as body and
// trailers is nil.
func Parse(message string) (body string, trailers []Trailer) {
	msg := strings.TrimRight(message, "\n")
	paragraphs := strings.Split(msg, "\n\n")
	// A commit message with only one paragraph is body-only — its subject
	// line ("fix: bug") often matches the trailer-line shape but is not a
	// trailer block. A trailer block must be a separate paragraph.
	if len(paragraphs) < 2 {
		return msg, nil
	}
	last := paragraphs[len(paragraphs)-1]
	lines := strings.Split(last, "\n")

	parsed := make([]Trailer, 0, len(lines))
	for _, ln := range lines {
		switch {
		case trailerLine.MatchString(ln):
			key, val, _ := strings.Cut(ln, ":")
			parsed = append(parsed, Trailer{
				Key:   key,
				Value: strings.TrimSpace(val),
			})
		case (strings.HasPrefix(ln, " ") || strings.HasPrefix(ln, "\t")) && len(parsed) > 0:
			parsed[len(parsed)-1].Value += " " + strings.TrimSpace(ln)
		default:
			return msg, nil
		}
	}
	if len(parsed) == 0 {
		return msg, nil
	}
	body = strings.Join(paragraphs[:len(paragraphs)-1], "\n\n")
	return body, parsed
}

// Format reassembles a commit message from a body and trailers. The body
// and trailer block are separated by exactly one blank line; the result
// always ends with a single trailing newline.
func Format(body string, trailers []Trailer) string {
	body = strings.TrimRight(body, "\n")
	if len(trailers) == 0 {
		if body == "" {
			return "\n"
		}
		return body + "\n"
	}
	var sb strings.Builder
	if body != "" {
		sb.WriteString(body)
		sb.WriteString("\n\n")
	}
	for _, t := range trailers {
		sb.WriteString(t.Key)
		sb.WriteString(": ")
		sb.WriteString(t.Value)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// Without returns ts with all entries whose Key (case-insensitive)
// matches one of keys removed.
func Without(ts []Trailer, keys []string) []Trailer {
	if len(keys) == 0 {
		return ts
	}
	drop := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		drop[strings.ToLower(k)] = struct{}{}
	}
	out := make([]Trailer, 0, len(ts))
	for _, t := range ts {
		if _, skip := drop[strings.ToLower(t.Key)]; !skip {
			out = append(out, t)
		}
	}
	return out
}
