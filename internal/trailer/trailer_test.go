package trailer

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name     string
		message  string
		body     string
		trailers []Trailer
	}{
		{
			name:    "no trailers",
			message: "fix: a bug\n",
			body:    "fix: a bug",
		},
		{
			name:    "empty",
			message: "",
			body:    "",
		},
		{
			name:    "single line subject",
			message: "first\n",
			body:    "first",
		},
		{
			name:    "subject + trailers",
			message: "subject\n\nReason: required\nTicket: JIRA-1\n",
			body:    "subject",
			trailers: []Trailer{
				{Key: "Reason", Value: "required"},
				{Key: "Ticket", Value: "JIRA-1"},
			},
		},
		{
			name:    "subject + body + trailers",
			message: "subject\n\nbody line one\nbody line two\n\nReason: r\nTicket: t\n",
			body:    "subject\n\nbody line one\nbody line two",
			trailers: []Trailer{
				{Key: "Reason", Value: "r"},
				{Key: "Ticket", Value: "t"},
			},
		},
		{
			name:    "folded value",
			message: "subj\n\nReason: line one\n  continuation\n",
			body:    "subj",
			trailers: []Trailer{
				{Key: "Reason", Value: "line one continuation"},
			},
		},
		{
			name:    "last paragraph not trailer-only is body",
			message: "subj\n\nThis paragraph: has a colon but isn't trailers.\nFree-form text here.\n",
			body:    "subj\n\nThis paragraph: has a colon but isn't trailers.\nFree-form text here.",
		},
		{
			name:    "preserves duplicates and order",
			message: "subj\n\nCo-authored-by: A <a@x>\nCo-authored-by: B <b@x>\n",
			body:    "subj",
			trailers: []Trailer{
				{Key: "Co-authored-by", Value: "A <a@x>"},
				{Key: "Co-authored-by", Value: "B <b@x>"},
			},
		},
		{
			name:    "empty trailer value",
			message: "subj\n\nKey:\n",
			body:    "subj",
			trailers: []Trailer{
				{Key: "Key", Value: ""},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body, trailers := Parse(c.message)
			if body != c.body {
				t.Errorf("body: got %q, want %q", body, c.body)
			}
			if !reflect.DeepEqual(trailers, c.trailers) {
				t.Errorf("trailers: got %#v, want %#v", trailers, c.trailers)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		trailers []Trailer
		want     string
	}{
		{name: "empty", want: "\n"},
		{name: "body only", body: "subj", want: "subj\n"},
		{
			name: "body + trailers",
			body: "subj",
			trailers: []Trailer{
				{Key: "Reason", Value: "r"},
				{Key: "Ticket", Value: "t"},
			},
			want: "subj\n\nReason: r\nTicket: t\n",
		},
		{
			name: "trailers only",
			trailers: []Trailer{
				{Key: "Reason", Value: "r"},
			},
			want: "Reason: r\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Format(c.body, c.trailers)
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	cases := []string{
		"subj\n",
		"subj\n\nbody\n\nReason: r\nTicket: t\n",
		"subj\n\nReason: r\n",
	}
	for _, msg := range cases {
		body, trailers := Parse(msg)
		got := Format(body, trailers)
		if got != msg {
			t.Errorf("round-trip mismatch:\n  in:  %q\n  out: %q", msg, got)
		}
	}
}

func TestWithout(t *testing.T) {
	ts := []Trailer{
		{Key: "Reason", Value: "r"},
		{Key: "Ticket", Value: "t"},
		{Key: "reason", Value: "r2"},
		{Key: "Foo", Value: "bar"},
	}
	got := Without(ts, []string{"Reason"})
	want := []Trailer{
		{Key: "Ticket", Value: "t"},
		{Key: "Foo", Value: "bar"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}

	if !reflect.DeepEqual(Without(ts, nil), ts) {
		t.Errorf("Without(_, nil) should return input unchanged")
	}
}
