package cli

import "strings"

// splitArgs separates positional arguments from flag arguments, allowing
// the user to write `gitmd add HEAD --reason r` as well as
// `gitmd add --reason r HEAD`. The stdlib flag package stops at the first
// non-flag, so we re-order positionals to the end.
//
// valueFlags lists long flag names whose value is the *next* argv element
// (e.g. "reason"). The list is needed so that "--reason demo" is recognized
// as a single flag invocation and "demo" doesn't end up classified as a
// positional. Boolean flags must be omitted.
//
// `--name=value` and short flags are passed through unchanged. A literal
// "--" terminates flag processing; everything after is positional.
func splitArgs(args []string, valueFlags map[string]struct{}) (flags, positional []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			positional = append(positional, args[i+1:]...)
			return
		}
		if !strings.HasPrefix(a, "-") {
			positional = append(positional, a)
			continue
		}
		flags = append(flags, a)
		if !strings.HasPrefix(a, "--") || strings.Contains(a, "=") {
			continue
		}
		name := strings.TrimPrefix(a, "--")
		if _, ok := valueFlags[name]; ok && i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}
	return
}
