package cli

import (
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	value := map[string]struct{}{"reason": {}, "trailer": {}}
	cases := []struct {
		name           string
		in             []string
		flags, posArgs []string
	}{
		{
			name:    "ref then flags",
			in:      []string{"HEAD", "--reason", "r", "--trailer", "K=V"},
			flags:   []string{"--reason", "r", "--trailer", "K=V"},
			posArgs: []string{"HEAD"},
		},
		{
			name:    "flags then ref",
			in:      []string{"--reason", "r", "HEAD"},
			flags:   []string{"--reason", "r"},
			posArgs: []string{"HEAD"},
		},
		{
			name:    "interspersed",
			in:      []string{"--reason", "r", "HEAD", "--trailer", "K=V"},
			flags:   []string{"--reason", "r", "--trailer", "K=V"},
			posArgs: []string{"HEAD"},
		},
		{
			name:    "equals form",
			in:      []string{"--reason=r", "HEAD"},
			flags:   []string{"--reason=r"},
			posArgs: []string{"HEAD"},
		},
		{
			name:    "double dash",
			in:      []string{"--reason", "r", "--", "-weird-ref"},
			flags:   []string{"--reason", "r"},
			posArgs: []string{"-weird-ref"},
		},
		{
			name:    "no positional",
			in:      []string{"--reason", "r"},
			flags:   []string{"--reason", "r"},
			posArgs: nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotFlags, gotPos := splitArgs(c.in, value)
			if !reflect.DeepEqual(gotFlags, c.flags) {
				t.Errorf("flags: got %#v, want %#v", gotFlags, c.flags)
			}
			if !reflect.DeepEqual(gotPos, c.posArgs) {
				t.Errorf("positional: got %#v, want %#v", gotPos, c.posArgs)
			}
		})
	}
}
