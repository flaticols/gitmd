package cli

import (
	"fmt"
	"os"

	"github.com/flaticols/gitmd/internal/git"
)

// runSeqEdit is the hidden `__seq-edit` subcommand. gitmd invokes itself
// as GIT_SEQUENCE_EDITOR during a non-HEAD amend; git then calls us with
// the path to the rebase todo file as the only positional argument.
//
// Required environment:
//
//	GITMD_TARGET   — full SHA of the commit whose message we're rewriting
//	GITMD_MSGFILE  — path to a file containing the new commit message
func runSeqEdit(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "gitmd __seq-edit: expected exactly one argument (todo file path)")
		return 2
	}
	target := os.Getenv("GITMD_TARGET")
	msgFile := os.Getenv("GITMD_MSGFILE")
	if target == "" || msgFile == "" {
		fmt.Fprintln(os.Stderr, "gitmd __seq-edit: GITMD_TARGET and GITMD_MSGFILE must be set")
		return 2
	}
	if err := git.RewriteRebaseTodo(args[0], target, msgFile); err != nil {
		fmt.Fprintf(os.Stderr, "gitmd __seq-edit: %s\n", err)
		return 1
	}
	return 0
}
