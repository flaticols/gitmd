// Package git wraps the git CLI. It is the only package in gitmd that
// invokes external commands; everything else operates on plain strings and
// structs and is unit-testable without git installed.
package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Error is returned for any non-zero exit from git. It carries the args
// invoked, captured stderr, and the underlying *exec.ExitError so callers
// can inspect the exit code via errors.As.
type Error struct {
	Err    error
	Stderr string
	Args   []string
}

func (e *Error) Error() string {
	cmd := "git " + strings.Join(e.Args, " ")
	if e.Stderr != "" {
		return fmt.Sprintf("%s: %s", cmd, e.Stderr)
	}
	return fmt.Sprintf("%s: %v", cmd, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

// ExitCode returns the underlying git process exit code, or -1 if the
// error didn't come from a process exit.
func (e *Error) ExitCode() int {
	if ee, ok := errors.AsType[*exec.ExitError](e.Err); ok {
		return ee.ExitCode()
	}
	return -1
}

func wrap(args []string, stderr string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{Err: err, Stderr: strings.TrimSpace(stderr), Args: args}
}

// Run invokes git and returns its stdout. On non-zero exit it returns *Error.
func Run(ctx context.Context, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", wrap(args, stderr.String(), err)
	}
	return stdout.String(), nil
}

// RunInput is like Run but pipes in to git's stdin.
func RunInput(ctx context.Context, in string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stdin = strings.NewReader(in)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", wrap(args, stderr.String(), err)
	}
	return stdout.String(), nil
}

// RunInherit invokes git with stdin/stdout/stderr connected to the user's
// terminal — used for rebase, where git may print progress and prompts.
// extraEnv entries (KEY=VAL) are appended to os.Environ().
func RunInherit(ctx context.Context, extraEnv []string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return wrap(args, "", err)
	}
	return nil
}
