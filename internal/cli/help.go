package cli

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// helpEntry describes a single command or flag for the help text.
type helpEntry struct{ name, desc string }

func writeRows(w io.Writer, rows []helpEntry) {
	tw := tabwriter.NewWriter(w, 2, 2, 2, ' ', 0)
	for _, r := range rows {
		fmt.Fprintf(tw, "  %s\t%s\n", r.name, r.desc)
	}
	tw.Flush()
}

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, "gitmd — manage git commit trailers (Reason, Ticket, Assisted, …)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  gitmd <command> [flags] [<ref>]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	writeRows(w, []helpEntry{
		{"add  [<ref>]", "append trailers to a commit"},
		{"set  [<ref>]", "replace trailers (any existing trailer with the same key is overwritten)"},
		{"del  [<ref>]", "remove trailers by key"},
		{"show [<ref>]", "pretty-print trailers"},
		{"version", "print gitmd and detected git version"},
		{"help [<cmd>]", "show help for a command"},
	})
	fmt.Fprintln(w)
	fmt.Fprintln(w, "<ref> defaults to HEAD for all commands.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Notes:")
	fmt.Fprintln(w, "  - Requires git >= 2.54.0 (uses git interpret-trailers and rebase --trailer machinery).")
	fmt.Fprintln(w, "  - add/set/del rewrite history. The working tree must be clean.")
	fmt.Fprintln(w, "  - <ref> is anything `git rev-parse` accepts (HEAD, HEAD~3, branch, short SHA).")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run `gitmd help <command>` for command-specific flags and examples.")
}

func printAddSetHelp(w io.Writer, cmd string) {
	verb := "Append"
	policy := "Skips trailers whose exact key/value pair is already present."
	if cmd == "set" {
		verb = "Replace"
		policy = "Any existing trailer with the same key is overwritten."
	}
	fmt.Fprintf(w, "gitmd %s — %s trailers on a commit.\n\n", cmd, verb)
	fmt.Fprintf(w, "Usage:\n  gitmd %s [<ref>] [flags]   (<ref> defaults to HEAD)\n\n", cmd)
	fmt.Fprintln(w, "Flags:")
	writeRows(w, []helpEntry{
		{"--reason VALUE", "shortcut for `Reason: VALUE`"},
		{"--ticket VALUE", "shortcut for `Ticket: VALUE`"},
		{"--assisted VALUE", "shortcut for `Assisted: VALUE`"},
		{"--trailer KEY=VALUE", "arbitrary trailer; repeatable"},
	})
	fmt.Fprintln(w)
	fmt.Fprintln(w, policy)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintf(w, "  gitmd %s HEAD --reason \"required for audit\" --ticket JIRA-1234\n", cmd)
	fmt.Fprintf(w, "  gitmd %s HEAD~2 --assisted \"Claude Code 4.7\" --trailer \"Reviewed-by=Alice <a@x>\"\n", cmd)
}

func printDelHelp(w io.Writer) {
	fmt.Fprintln(w, "gitmd del — remove trailers from a commit by key.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:\n  gitmd del [<ref>] [flags]   (<ref> defaults to HEAD)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	writeRows(w, []helpEntry{
		{"--reason", "remove `Reason` trailer(s)"},
		{"--ticket", "remove `Ticket` trailer(s)"},
		{"--assisted", "remove `Assisted` trailer(s)"},
		{"--key NAME", "remove trailer with the given key; repeatable"},
	})
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Key matching is case-insensitive. All trailers with a matching key are removed.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  gitmd del HEAD --reason --ticket")
	fmt.Fprintln(w, "  gitmd del HEAD~3 --key Reviewed-by --key Tested-by")
}

func printShowHelp(w io.Writer) {
	fmt.Fprintln(w, "gitmd show — print the trailers of a commit.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:\n  gitmd show [<ref>] [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	writeRows(w, []helpEntry{
		{"--json", "machine-readable output"},
	})
	fmt.Fprintln(w)
	fmt.Fprintln(w, "<ref> defaults to HEAD. Pretty output uses ANSI on a TTY (honors NO_COLOR).")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  gitmd show")
	fmt.Fprintln(w, "  gitmd show HEAD~3")
	fmt.Fprintln(w, "  gitmd show HEAD --json | jq .trailers")
}

// runHelp dispatches `gitmd help [<command>]`.
func runHelp(args []string) int {
	if len(args) == 0 {
		printRootHelp(os.Stdout)
		return 0
	}
	switch args[0] {
	case "add":
		printAddSetHelp(os.Stdout, "add")
	case "set":
		printAddSetHelp(os.Stdout, "set")
	case "del":
		printDelHelp(os.Stdout)
	case "show":
		printShowHelp(os.Stdout)
	case "version":
		fmt.Fprintln(os.Stdout, "gitmd version — print gitmd and detected git version.")
	case "help":
		fmt.Fprintln(os.Stdout, "gitmd help [<command>] — show help.")
	default:
		fmt.Fprintf(os.Stderr, "gitmd help: unknown command %q\n", args[0])
		return 2
	}
	return 0
}

// suppressFlagOutput silences the default flag package "usage:" output so
// we can render our own help on parse errors.
func suppressFlagOutput(w io.Writer) io.Writer {
	if w == nil {
		return io.Discard
	}
	return w
}
