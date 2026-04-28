// Package render formats commit views for human and machine consumption.
package render

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"golang.org/x/term"

	"github.com/flaticols/gitmd/internal/trailer"
)

// View is the canonical shape we render — both pretty and JSON sinks
// consume it.
type View struct {
	SHA      string            `json:"commit"`
	Subject  string            `json:"subject"`
	Author   Author            `json:"author"`
	Trailers []trailer.Trailer `json:"trailers"`
}

// Author identifies the author for both pretty and JSON output.
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

// MarshalJSON gives Trailer a stable on-the-wire shape ({"key","value"})
// independent of the Go field names.
func trailerJSON(t trailer.Trailer) any {
	return map[string]string{"key": t.Key, "value": t.Value}
}

const (
	ansiReset = "\x1b[0m"
	ansiDim   = "\x1b[2m"
	ansiBold  = "\x1b[1m"
)

// UseColor reports whether ANSI escapes should be emitted on f. Honors
// the NO_COLOR environment variable and falls back to TTY detection.
func UseColor(f *os.File) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// Pretty writes a coloured, aligned commit view to w.
func Pretty(w io.Writer, color bool, v View) error {
	dim, bold, reset := "", "", ""
	if color {
		dim, bold, reset = ansiDim, ansiBold, ansiReset
	}
	short := v.SHA
	if len(short) > 8 {
		short = short[:8]
	}
	if _, err := fmt.Fprintf(w, "%scommit%s  %s — %s\n", dim, reset, short, v.Subject); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%sauthor%s  %s <%s>  %s\n",
		dim, reset, v.Author.Name, v.Author.Email, v.Author.Date); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if len(v.Trailers) == 0 {
		_, err := fmt.Fprintf(w, "%s(no trailers)%s\n", dim, reset)
		return err
	}
	tw := tabwriter.NewWriter(w, 2, 2, 2, ' ', 0)
	for _, t := range v.Trailers {
		if _, err := fmt.Fprintf(tw, "  %s%s%s\t%s\n", bold, t.Key, reset, t.Value); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// JSON writes v as indented JSON.
func JSON(w io.Writer, v View) error {
	out := struct {
		SHA      string `json:"commit"`
		Subject  string `json:"subject"`
		Author   Author `json:"author"`
		Trailers []any  `json:"trailers"`
	}{
		SHA:     v.SHA,
		Subject: v.Subject,
		Author:  v.Author,
	}
	out.Trailers = make([]any, 0, len(v.Trailers))
	for _, t := range v.Trailers {
		out.Trailers = append(out.Trailers, trailerJSON(t))
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
