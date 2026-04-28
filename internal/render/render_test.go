package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/flaticols/gitmd/internal/trailer"
)

func sample() View {
	return View{
		SHA:     "a1b2c3d4e5f6789012345678901234567890abcd",
		Subject: "Add login flow",
		Author: Author{
			Name:  "Denis Panfilov",
			Email: "services@flaticols.dev",
			Date:  "2026-04-28T10:00:00+02:00",
		},
		Trailers: []trailer.Trailer{
			{Key: "Reason", Value: "Required for security audit"},
			{Key: "Ticket", Value: "JIRA-1234"},
		},
	}
}

func TestPretty_NoColor(t *testing.T) {
	var buf bytes.Buffer
	if err := Pretty(&buf, false, sample()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "\x1b[") {
		t.Errorf("expected no ANSI escapes when color=false, got:\n%s", out)
	}
	for _, want := range []string{"commit  a1b2c3d4 — Add login flow", "Reason", "JIRA-1234"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPretty_Color(t *testing.T) {
	var buf bytes.Buffer
	if err := Pretty(&buf, true, sample()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "\x1b[") {
		t.Error("expected ANSI escapes when color=true")
	}
}

func TestPretty_NoTrailers(t *testing.T) {
	v := sample()
	v.Trailers = nil
	var buf bytes.Buffer
	if err := Pretty(&buf, false, v); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "(no trailers)") {
		t.Errorf("expected '(no trailers)' marker, got:\n%s", buf.String())
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, sample()); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		`"commit": "a1b2c3d4e5f6789012345678901234567890abcd"`,
		`"subject": "Add login flow"`,
		`"key": "Reason"`,
		`"value": "JIRA-1234"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("JSON output missing %q:\n%s", want, got)
		}
	}
}
