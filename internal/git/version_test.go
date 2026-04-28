package git

import (
	"testing"

	"github.com/flaticols/server"
)

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in            string
		major, minor  int
		patch         int
		expectInvalid bool
	}{
		{in: "git version 2.54.0", major: 2, minor: 54, patch: 0},
		{in: "git version 2.54.0.windows.1", major: 2, minor: 54, patch: 0},
		{in: "git version 2.54.0-rc1", major: 2, minor: 54, patch: 0},
		{in: "git version 3.0.0", major: 3, minor: 0, patch: 0},
		{in: "git version 2.53.1", major: 2, minor: 53, patch: 1},
		{in: "not a version", expectInvalid: true},
	}
	for _, c := range cases {
		got, err := ParseVersion(c.in)
		if c.expectInvalid {
			if err == nil {
				t.Errorf("ParseVersion(%q) should have errored", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseVersion(%q) error: %v", c.in, err)
			continue
		}
		if got.Major != c.major || got.Minor != c.minor || got.Patch != c.patch {
			t.Errorf("ParseVersion(%q) = %d.%d.%d, want %d.%d.%d",
				c.in, got.Major, got.Minor, got.Patch, c.major, c.minor, c.patch)
		}
	}
}

func TestVersionMinComparison(t *testing.T) {
	min := server.New(2, 54, 0, nil, nil)
	cases := []struct {
		v        server.Version
		atLeast  bool
		describe string
	}{
		{server.New(2, 54, 0, nil, nil), true, "exact"},
		{server.New(2, 54, 1, nil, nil), true, "patch newer"},
		{server.New(2, 55, 0, nil, nil), true, "minor newer"},
		{server.New(3, 0, 0, nil, nil), true, "major newer"},
		{server.New(2, 53, 99, nil, nil), false, "minor older"},
		{server.New(1, 99, 99, nil, nil), false, "major older"},
	}
	for _, c := range cases {
		got := c.v.GreaterThanOrEqual(min)
		if got != c.atLeast {
			t.Errorf("(%s).GreaterThanOrEqual(%s) [%s] = %v, want %v",
				c.v, min, c.describe, got, c.atLeast)
		}
	}
}
