package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Commit holds the metadata gitmd needs about a single commit.
type Commit struct {
	SHA     string
	Subject string
	Author  string
	Email   string
	Date    string
	Message string
	Parents []string
}

// IsMerge reports whether the commit has more than one parent.
func (c Commit) IsMerge() bool { return len(c.Parents) > 1 }

// commitFormat keeps fields in a fixed order so we can SplitN them out.
// %B is last because the body may contain anything, including newlines.
const commitFormat = "%H%n%s%n%an%n%ae%n%aI%n%P%n%B"

// Resolve returns the full SHA for ref. ref must point to a commit.
func Resolve(ctx context.Context, ref string) (string, error) {
	out, err := Run(ctx, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Head returns the full SHA of HEAD.
func Head(ctx context.Context) (string, error) {
	return Resolve(ctx, "HEAD")
}

// ReadCommit fetches the commit metadata for sha.
func ReadCommit(ctx context.Context, sha string) (Commit, error) {
	out, err := Run(ctx, "log", "-1", "--format="+commitFormat, sha)
	if err != nil {
		return Commit{}, err
	}
	parts := strings.SplitN(out, "\n", 7)
	if len(parts) < 7 {
		return Commit{}, fmt.Errorf("unexpected git log output: %q", out)
	}
	return Commit{
		SHA:     parts[0],
		Subject: parts[1],
		Author:  parts[2],
		Email:   parts[3],
		Date:    parts[4],
		Parents: strings.Fields(parts[5]),
		Message: strings.TrimRight(parts[6], "\n"),
	}, nil
}

// IsClean reports whether the working tree has no untracked, unstaged, or
// staged changes.
func IsClean(ctx context.Context) (bool, error) {
	out, err := Run(ctx, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "", nil
}

// IsAncestorOfHEAD reports whether sha is reachable from HEAD.
func IsAncestorOfHEAD(ctx context.Context, sha string) (bool, error) {
	_, err := Run(ctx, "merge-base", "--is-ancestor", sha, "HEAD")
	if err == nil {
		return true, nil
	}
	if ge, ok := errors.AsType[*Error](err); ok && ge.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

// AmendHEADMessage replaces HEAD's commit message with msg via
// `git commit --amend -F <tmpfile>`.
func AmendHEADMessage(ctx context.Context, msg string) error {
	path, cleanup, err := writeTempMessage(msg)
	if err != nil {
		return err
	}
	defer cleanup()
	_, err = Run(ctx, "commit", "--amend", "--no-edit", "--allow-empty", "-F", path)
	return err
}

// InterpretTrailers pipes msg through `git interpret-trailers` with the
// given extra args. Used to compose new trailer blocks; we never hand-format
// them ourselves.
func InterpretTrailers(ctx context.Context, msg string, extraArgs ...string) (string, error) {
	args := append([]string{"interpret-trailers"}, extraArgs...)
	return RunInput(ctx, msg, args...)
}

// writeTempMessage writes msg to a temp file and returns its path along with
// a cleanup function that removes it.
func writeTempMessage(msg string) (string, func(), error) {
	f, err := os.CreateTemp("", "gitmd-msg-*")
	if err != nil {
		return "", nil, err
	}
	if _, err := f.WriteString(msg); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", nil, err
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", nil, err
	}
	return f.Name(), func() { os.Remove(f.Name()) }, nil
}
