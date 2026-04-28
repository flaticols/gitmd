package git

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/flaticols/server"
)

// MinVersion is the minimum git version gitmd supports — `git rebase
// --trailer` from 2.54 is the headline reason this tool exists.
var MinVersion = server.New(2, 54, 0, nil, nil)

// gitVersionRE captures the X.Y[.Z][-rcN][.windows.N] tail of `git --version`
// output. We want everything from the first digit onward; flaticols/server
// then takes care of SemVer parsing (with one bit of preprocessing for the
// non-SemVer ".windows.N" suffix git ships on Windows).
var gitVersionRE = regexp.MustCompile(`(\d+(?:\.\d+){1,3}(?:[-.][0-9A-Za-z][\w.-]*)?)`)

// ParseVersion extracts a server.Version from output like
// "git version 2.54.0", "git version 2.54.0.windows.1", or
// "git version 2.54.0-rc1".
func ParseVersion(s string) (server.Version, error) {
	m := gitVersionRE.FindString(strings.TrimSpace(s))
	if m == "" {
		return server.Version{}, fmt.Errorf("cannot parse git version: %q", s)
	}
	// `2.54.0.windows.1` is not SemVer (four dotted components); drop the
	// platform suffix and treat it as build metadata so it survives in
	// String() if anyone prints it back.
	parts := strings.SplitN(m, ".", 4)
	var meta []string
	if len(parts) == 4 {
		meta = strings.Split(parts[3], ".")
		m = strings.Join(parts[:3], ".")
	}
	v, err := server.Parse(m)
	if err != nil {
		return server.Version{}, fmt.Errorf("parse git version %q: %w", s, err)
	}
	if len(meta) > 0 {
		v = server.SetMetadata(v, meta)
	}
	return v, nil
}

// DetectVersion runs `git --version` and parses the result.
func DetectVersion(ctx context.Context) (server.Version, error) {
	out, err := Run(ctx, "--version")
	if err != nil {
		return server.Version{}, err
	}
	return ParseVersion(out)
}

// RequireMinVersion fails if the installed git is older than MinVersion.
func RequireMinVersion(ctx context.Context) (server.Version, error) {
	v, err := DetectVersion(ctx)
	if err != nil {
		return v, err
	}
	if v.LessThan(MinVersion) {
		return v, fmt.Errorf("gitmd requires git >= %s; found %s", MinVersion, v)
	}
	return v, nil
}
