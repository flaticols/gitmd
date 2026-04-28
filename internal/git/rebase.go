package git

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// AmendNonHEADMessage rewrites the commit message of a non-HEAD commit by
// running `git rebase -i <sha>^` with this binary as GIT_SEQUENCE_EDITOR.
// The sequence editor (registered as the hidden `__seq-edit` subcommand)
// reads the desired message file path from GITMD_MSGFILE and the target
// SHA from GITMD_TARGET, then injects an `exec git commit --amend -F
// <msgfile>` immediately after the target's pick line in the rebase todo.
//
// This is more surgical than `git rebase --trailer`, which applies trailers
// to every commit in the rebased range.
func AmendNonHEADMessage(ctx context.Context, sha, msg string) error {
	msgPath, cleanup, err := writeTempMessage(msg)
	if err != nil {
		return err
	}
	defer cleanup()

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate gitmd binary: %w", err)
	}

	root, err := isRootCommit(ctx, sha)
	if err != nil {
		return err
	}
	rebaseArgs := []string{"rebase", "-i"}
	if root {
		rebaseArgs = append(rebaseArgs, "--root")
	} else {
		rebaseArgs = append(rebaseArgs, sha+"^")
	}

	env := []string{
		"GITMD_TARGET=" + sha,
		"GITMD_MSGFILE=" + msgPath,
		"GIT_SEQUENCE_EDITOR=" + shellSingleQuote(self) + " __seq-edit",
		"GIT_EDITOR=true",
	}
	return RunInherit(ctx, env, rebaseArgs...)
}

func isRootCommit(ctx context.Context, sha string) (bool, error) {
	out, err := Run(ctx, "rev-list", "--parents", "-n", "1", sha)
	if err != nil {
		return false, err
	}
	return len(strings.Fields(out)) == 1, nil
}

// shellSingleQuote single-quotes s for safe inclusion in a POSIX shell
// command (git invokes GIT_SEQUENCE_EDITOR via the shell).
func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// RewriteRebaseTodo edits an `git rebase -i` todo file in place, inserting
// `exec git commit --amend --no-edit -F <msgPath>` immediately after the
// pick line for targetSHA. It is invoked from the `__seq-edit` subcommand
// in the cli package; it lives here so the shell-quoting rule is co-located
// with the outer rebase orchestration.
func RewriteRebaseTodo(todoPath, targetSHA, msgPath string) error {
	data, err := os.ReadFile(todoPath)
	if err != nil {
		return err
	}
	exec := "exec git commit --amend --no-edit -F " + shellSingleQuote(msgPath)

	var out strings.Builder
	inserted := false
	for line := range strings.SplitSeq(string(data), "\n") {
		out.WriteString(line)
		out.WriteByte('\n')
		if !inserted && pickMatches(line, targetSHA) {
			out.WriteString(exec)
			out.WriteByte('\n')
			inserted = true
		}
	}
	if !inserted {
		return fmt.Errorf("target commit %s not found in rebase todo %s", targetSHA[:7], todoPath)
	}
	// Trim the extra trailing newline added by the loop.
	result := strings.TrimRight(out.String(), "\n") + "\n"
	return os.WriteFile(todoPath, []byte(result), 0o644)
}

// pickMatches reports whether line is a `pick <sha> ...` instruction whose
// SHA is a prefix of fullSHA.
func pickMatches(line, fullSHA string) bool {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return false
	}
	if fields[0] != "pick" && fields[0] != "p" {
		return false
	}
	return strings.HasPrefix(fullSHA, fields[1])
}
