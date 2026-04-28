# gitmd

A small Go CLI for managing **git commit trailers** — the structured
`Key: Value` lines at the end of commit messages (`Signed-off-by:`,
`Reviewed-by:`, …). Built on top of git ≥ 2.54's
`interpret-trailers` and `rebase --trailer` machinery.

## Install

### Go (system-wide)

```sh
go install github.com/flaticols/gitmd/cmd/gitmd@latest
```

### Go tool dependency (per-project, Go 1.24+)

Pin gitmd as a project-local tool — version recorded in `go.mod`, runnable
via `go tool`:

```sh
go get -tool github.com/flaticols/gitmd/cmd/gitmd@latest
go tool gitmd show
```

### Nix (macOS)

The release pipeline auto-publishes a Nix derivation to the `latest` branch
of this repo on every tag. Install ad-hoc:

```sh
nix-env -if https://github.com/flaticols/gitmd/raw/latest/nix/gitmd.nix
```

Or pin it in a flake:

```nix
gitmd = pkgs.callPackage (builtins.fetchurl
  "https://github.com/flaticols/gitmd/raw/latest/nix/gitmd.nix") {};
```

Currently ships `aarch64-darwin` and `x86_64-darwin` binaries.

## Quick start

```sh
# operate on HEAD (default)
gitmd add --reason "required for audit" --ticket JIRA-1234
gitmd set --ticket JIRA-2
gitmd del --reason
gitmd show

# explicit ref — anything `git rev-parse` accepts
gitmd add HEAD~2 --assisted "Claude Code 4.7" --trailer "Reviewed-by=Alice <a@x>"
gitmd show HEAD~3 --json
```

### What `show` looks like

Pretty form — what you read:

```
commit  33d2dff8 — docs: add retro-styled single-page HTML reference manual
author  Denis Panfilov <gh@flaticols.dev>  2026-04-28T20:21:52+02:00

  Reason    single-page reference manual for the project; satisfies the docs/ workflow added on origin
  Assisted  Claude Code (Opus 4.7)
```

JSON form — what your scripts read:

```json
{
  "commit": "33d2dff8be6ea9b3a534d64cc3e13773cf1cc439",
  "subject": "docs: add retro-styled single-page HTML reference manual",
  "author": {
    "name": "Denis Panfilov",
    "email": "gh@flaticols.dev",
    "date": "2026-04-28T20:21:52+02:00"
  },
  "trailers": [
    { "key": "Reason",   "value": "single-page reference manual for the project; satisfies the docs/ workflow added on origin" },
    { "key": "Assisted", "value": "Claude Code (Opus 4.7)" }
  ]
}
```

## Commands

`<ref>` is optional everywhere and defaults to `HEAD`.

| Command          | Purpose                                                     |
|------------------|-------------------------------------------------------------|
| `add  [<ref>]`   | append trailers (idempotent — skips identical neighbours)   |
| `set  [<ref>]`   | replace trailers with the same key                          |
| `del  [<ref>]`   | remove trailers by key                                      |
| `show [<ref>]`   | pretty-print trailers (`--json` for machines)               |
| `version`        | print gitmd + detected git version                          |

### Predefined shortcuts

`--reason VALUE`, `--ticket VALUE`, `--assisted VALUE` map to canonical
keys `Reason`, `Ticket`, `Assisted`. Use `--trailer KEY=VALUE` (repeatable)
for anything else. `<ref>` may appear before or after the flags.

## What it does

- For `HEAD`, gitmd shells out to `git commit --amend`.
- For any other reachable commit, gitmd runs `git rebase -i <ref>^` with
  itself registered as `GIT_SEQUENCE_EDITOR`. The sequence editor inserts a
  surgical `exec git commit --amend -F <newmsg>` after the target's
  `pick` line — so only the target commit is rewritten (every descendant
  is rebased on top, but its message is untouched).
- Trailer composition is delegated entirely to `git interpret-trailers`,
  so gitmd inherits git's exact semantics (folded values, `---` divider,
  the 25 % trailer-line tolerance).

## Safety

- Refuses to run on a dirty working tree.
- Refuses to amend non-HEAD merge commits.
- Never force-pushes; if the rewritten commit was already published you
  must run `git push --force-with-lease` yourself.
- On rebase failure, surfaces git's error verbatim — recover with
  `git rebase --abort`.

## Build / test

```sh
make build   # binary in ./gitmd
make test
make lint
```

## License

[MIT](LICENSE)
