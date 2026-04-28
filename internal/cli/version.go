package cli

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/flaticols/gitmd/internal/git"
)

// gitmdVersion derives the version from build info (module main version,
// or VCS revision + dirty flag for `go install`/`go run` builds).
func gitmdVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	var rev, modified string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			if s.Value == "true" {
				modified = "-dirty"
			}
		}
	}
	if rev == "" {
		return "devel"
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	return strings.ToLower(rev) + modified
}

func runVersion() int {
	fmt.Printf("gitmd %s\n", gitmdVersion())
	v, err := git.DetectVersion(context.Background())
	if err != nil {
		fmt.Printf("git    (not detected: %v)\n", err)
		return 0
	}
	fmt.Printf("git    %s\n", v)
	return 0
}
