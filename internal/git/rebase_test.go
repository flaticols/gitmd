package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRewriteRebaseTodo(t *testing.T) {
	dir := t.TempDir()
	todo := filepath.Join(dir, "todo")
	msgFile := filepath.Join(dir, "msg")

	target := "a1b2c3d4e5f6789012345678901234567890abcd"
	original := strings.Join([]string{
		"pick 1111111 first commit",
		"pick a1b2c3d second commit",
		"pick 2222222 third commit",
		"",
		"# Rebase instructions...",
	}, "\n")
	if err := os.WriteFile(todo, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := RewriteRebaseTodo(todo, target, msgFile); err != nil {
		t.Fatalf("RewriteRebaseTodo: %v", err)
	}

	got, err := os.ReadFile(todo)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{
		"pick 1111111 first commit",
		"pick a1b2c3d second commit",
		"exec git commit --amend --no-edit -F '" + msgFile + "'",
		"pick 2222222 third commit",
		"",
		"# Rebase instructions...",
	}, "\n") + "\n"
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRewriteRebaseTodo_MissingTarget(t *testing.T) {
	dir := t.TempDir()
	todo := filepath.Join(dir, "todo")
	if err := os.WriteFile(todo, []byte("pick 1111111 first\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := RewriteRebaseTodo(todo, "9999999999999999999999999999999999999999", "/tmp/msg")
	if err == nil {
		t.Fatal("expected error when target not in todo")
	}
}

func TestShellSingleQuote(t *testing.T) {
	cases := map[string]string{
		"simple":      "'simple'",
		"with space":  "'with space'",
		"it's a path": `'it'\''s a path'`,
	}
	for in, want := range cases {
		if got := shellSingleQuote(in); got != want {
			t.Errorf("shellSingleQuote(%q) = %q, want %q", in, got, want)
		}
	}
}
