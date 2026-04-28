// Package cli implements the gitmd command-line interface.
//
// The entry point is Run, which dispatches on the first arg. Each
// subcommand owns a small file in this package; logic shared between
// add/set/del lives in apply.go.
package cli

import (
	"fmt"
	"io"
	"os"
)

// Run is the main entry point. It returns the process exit code.
func Run(args []string) int {
	if len(args) == 0 {
		printRootHelp(os.Stdout)
		return 0
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "help", "--help", "-h":
		return runHelp(rest)
	case "version", "--version":
		return runVersion()
	case "show":
		return runShow(rest)
	case "add":
		return runApply(opAdd, rest)
	case "set":
		return runApply(opSet, rest)
	case "del":
		return runDel(rest)
	case "__seq-edit":
		return runSeqEdit(rest)
	default:
		fmt.Fprintf(os.Stderr, "gitmd: unknown command %q\n\n", cmd)
		printRootHelp(os.Stderr)
		return 2
	}
}

// errExit prints err to stderr and returns the appropriate exit code.
func errExit(err error) int {
	if err == nil {
		return 0
	}
	fmt.Fprintf(os.Stderr, "gitmd: %s\n", err)
	return 1
}

// usageError prints a "usage" error and returns exit 2.
func usageError(w io.Writer, format string, args ...any) int {
	fmt.Fprintf(w, "gitmd: ")
	fmt.Fprintf(w, format, args...)
	fmt.Fprintln(w)
	return 2
}
