package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/flaticols/gitmd/internal/git"
	"github.com/flaticols/gitmd/internal/trailer"
)

// opKind selects between the two `git interpret-trailers --if-exists`
// policies used by the add and set subcommands.
type opKind int

const (
	opAdd opKind = iota // policy: addIfDifferentNeighbor (idempotent append)
	opSet               // policy: replace
)

func (o opKind) name() string {
	if o == opSet {
		return "set"
	}
	return "add"
}

func (o opKind) ifExists() string {
	if o == opSet {
		return "replace"
	}
	return "addIfDifferentNeighbor"
}

// trailerFlag is a repeated --trailer KEY=VALUE flag.
type trailerFlag []trailer.Trailer

func (f *trailerFlag) String() string {
	parts := make([]string, 0, len(*f))
	for _, t := range *f {
		parts = append(parts, t.Key+"="+t.Value)
	}
	return strings.Join(parts, ",")
}

func (f *trailerFlag) Set(s string) error {
	key, val, ok := strings.Cut(s, "=")
	if !ok {
		return fmt.Errorf(`expected "Key=Value", got %q`, s)
	}
	key = strings.TrimSpace(key)
	val = strings.TrimSpace(val)
	if key == "" {
		return fmt.Errorf("empty trailer key in %q", s)
	}
	*f = append(*f, trailer.Trailer{Key: key, Value: val})
	return nil
}

// stringList is a repeated string-valued flag (e.g. --key NAME).
type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

var addSetValueFlags = map[string]struct{}{
	"reason":   {},
	"ticket":   {},
	"assisted": {},
	"trailer":  {},
}

func runApply(op opKind, args []string) int {
	name := op.name()
	flagArgs, positional := splitArgs(args, addSetValueFlags)
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(suppressFlagOutput(nil))
	var (
		reason   string
		ticket   string
		assisted string
		extra    trailerFlag
	)
	fs.StringVar(&reason, "reason", "", "")
	fs.StringVar(&ticket, "ticket", "", "")
	fs.StringVar(&assisted, "assisted", "", "")
	fs.Var(&extra, "trailer", "")
	if err := fs.Parse(flagArgs); err != nil {
		printAddSetHelp(os.Stderr, name)
		return 2
	}
	if len(positional) > 1 {
		return usageError(os.Stderr, "%s: expected at most one <ref>, got %d", name, len(positional))
	}
	ref := "HEAD"
	if len(positional) == 1 {
		ref = positional[0]
	}

	trailers := collectTrailers(reason, ticket, assisted, extra)
	if len(trailers) == 0 {
		return usageError(os.Stderr, "%s: no trailers given (try --reason, --ticket, --assisted, or --trailer KEY=VALUE)", name)
	}

	ctx := context.Background()
	if err := preflight(ctx); err != nil {
		return errExit(err)
	}
	if err := applyTrailers(ctx, ref, op, trailers); err != nil {
		return errExit(err)
	}
	return 0
}

var delValueFlags = map[string]struct{}{"key": {}}

func runDel(args []string) int {
	flagArgs, positional := splitArgs(args, delValueFlags)
	fs := flag.NewFlagSet("del", flag.ContinueOnError)
	fs.SetOutput(suppressFlagOutput(nil))
	var (
		reason, ticket, assisted bool
		keys                     stringList
	)
	fs.BoolVar(&reason, "reason", false, "")
	fs.BoolVar(&ticket, "ticket", false, "")
	fs.BoolVar(&assisted, "assisted", false, "")
	fs.Var(&keys, "key", "")
	if err := fs.Parse(flagArgs); err != nil {
		printDelHelp(os.Stderr)
		return 2
	}
	if len(positional) > 1 {
		return usageError(os.Stderr, "del: expected at most one <ref>, got %d", len(positional))
	}
	ref := "HEAD"
	if len(positional) == 1 {
		ref = positional[0]
	}

	if reason {
		keys = append(keys, "Reason")
	}
	if ticket {
		keys = append(keys, "Ticket")
	}
	if assisted {
		keys = append(keys, "Assisted")
	}
	if len(keys) == 0 {
		return usageError(os.Stderr, "del: no keys given (try --reason, --ticket, --assisted, or --key NAME)")
	}

	ctx := context.Background()
	if err := preflight(ctx); err != nil {
		return errExit(err)
	}
	if err := deleteTrailers(ctx, ref, keys); err != nil {
		return errExit(err)
	}
	return 0
}

// collectTrailers merges the predefined-shortcut flags with --trailer
// entries, preserving the order users typically expect (Reason, Ticket,
// Assisted, then arbitrary).
func collectTrailers(reason, ticket, assisted string, extra []trailer.Trailer) []trailer.Trailer {
	out := make([]trailer.Trailer, 0, 3+len(extra))
	if reason != "" {
		out = append(out, trailer.Trailer{Key: "Reason", Value: reason})
	}
	if ticket != "" {
		out = append(out, trailer.Trailer{Key: "Ticket", Value: ticket})
	}
	if assisted != "" {
		out = append(out, trailer.Trailer{Key: "Assisted", Value: assisted})
	}
	out = append(out, extra...)
	return out
}

// preflight runs the checks required before any history-rewriting op:
// minimum git version and a clean working tree.
func preflight(ctx context.Context) error {
	if _, err := git.RequireMinVersion(ctx); err != nil {
		return err
	}
	clean, err := git.IsClean(ctx)
	if err != nil {
		return err
	}
	if !clean {
		return fmt.Errorf("working tree is not clean; commit or stash changes first")
	}
	return nil
}

// applyTrailers composes the new commit message via `git interpret-trailers`
// and writes it back to the target commit (HEAD or non-HEAD path).
func applyTrailers(ctx context.Context, ref string, op opKind, trailers []trailer.Trailer) error {
	sha, c, isHead, err := resolveTarget(ctx, ref)
	if err != nil {
		return err
	}
	if !isHead && c.IsMerge() {
		return fmt.Errorf("refusing to amend merge commit %s (rewriting merges across a rebase is unsupported)", sha[:7])
	}

	args := []string{
		"--if-exists=" + op.ifExists(),
		"--where=end",
		"--no-divider",
	}
	for _, t := range trailers {
		args = append(args, "--trailer", t.Key+": "+t.Value)
	}
	newMsg, err := git.InterpretTrailers(ctx, c.Message, args...)
	if err != nil {
		return err
	}
	// interpret-trailers may add a trailing newline that the source message
	// (read with TrimRight) lacks; compare with both sides normalized.
	if strings.TrimRight(newMsg, "\n") == strings.TrimRight(c.Message, "\n") {
		fmt.Fprintf(os.Stderr, "gitmd: %s: nothing to do (trailers already present)\n", op.name())
		return nil
	}
	return writeBack(ctx, sha, isHead, newMsg)
}

// deleteTrailers parses the existing commit message, drops any trailers
// whose key matches keys, and writes the result back.
func deleteTrailers(ctx context.Context, ref string, keys []string) error {
	sha, c, isHead, err := resolveTarget(ctx, ref)
	if err != nil {
		return err
	}
	if !isHead && c.IsMerge() {
		return fmt.Errorf("refusing to amend merge commit %s", sha[:7])
	}
	body, trailers := trailer.Parse(c.Message)
	remaining := trailer.Without(trailers, keys)
	if len(remaining) == len(trailers) {
		fmt.Fprintln(os.Stderr, "gitmd: del: nothing to do (no matching trailers)")
		return nil
	}
	newMsg := trailer.Format(body, remaining)
	return writeBack(ctx, sha, isHead, newMsg)
}

// resolveTarget verifies ref points at a reachable commit and returns its
// metadata. The boolean is true when the target is HEAD.
func resolveTarget(ctx context.Context, ref string) (sha string, c git.Commit, isHead bool, err error) {
	sha, err = git.Resolve(ctx, ref)
	if err != nil {
		return "", git.Commit{}, false, fmt.Errorf("resolve %q: %w", ref, err)
	}
	head, err := git.Head(ctx)
	if err != nil {
		return "", git.Commit{}, false, err
	}
	if sha != head {
		var ok bool
		ok, err = git.IsAncestorOfHEAD(ctx, sha)
		if err != nil {
			return "", git.Commit{}, false, err
		}
		if !ok {
			return "", git.Commit{}, false, fmt.Errorf("commit %s is not reachable from HEAD", sha[:7])
		}
	}
	c, err = git.ReadCommit(ctx, sha)
	if err != nil {
		return "", git.Commit{}, false, err
	}
	return sha, c, sha == head, nil
}

// writeBack updates the target commit's message in place. For HEAD this is
// a `git commit --amend`; for any other reachable commit we run a
// surgical interactive rebase.
func writeBack(ctx context.Context, sha string, isHead bool, newMsg string) error {
	if isHead {
		return git.AmendHEADMessage(ctx, newMsg)
	}
	if err := git.AmendNonHEADMessage(ctx, sha, newMsg); err != nil {
		return fmt.Errorf("%w\n(if a rebase is in progress, run `git rebase --abort` to recover)", err)
	}
	newHead, err := git.Head(ctx)
	if err == nil {
		fmt.Fprintf(os.Stderr, "gitmd: rewrote history; HEAD is now %s\n", newHead[:8])
	}
	return nil
}
