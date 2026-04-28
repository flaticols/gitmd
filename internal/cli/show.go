package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/flaticols/gitmd/internal/git"
	"github.com/flaticols/gitmd/internal/render"
	"github.com/flaticols/gitmd/internal/trailer"
)

func runShow(args []string) int {
	flagArgs, positional := splitArgs(args, nil)
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(suppressFlagOutput(nil))
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(flagArgs); err != nil {
		printShowHelp(os.Stderr)
		return 2
	}
	if len(positional) > 1 {
		return usageError(os.Stderr, "show: expected at most one <ref>, got %d", len(positional))
	}
	ref := "HEAD"
	if len(positional) == 1 {
		ref = positional[0]
	}

	ctx := context.Background()
	if _, err := git.RequireMinVersion(ctx); err != nil {
		return errExit(err)
	}

	sha, err := git.Resolve(ctx, ref)
	if err != nil {
		return errExit(fmt.Errorf("resolve %q: %w", ref, err))
	}
	c, err := git.ReadCommit(ctx, sha)
	if err != nil {
		return errExit(err)
	}
	_, trailers := trailer.Parse(c.Message)

	v := render.View{
		SHA:     c.SHA,
		Subject: c.Subject,
		Author: render.Author{
			Name:  c.Author,
			Email: c.Email,
			Date:  c.Date,
		},
		Trailers: trailers,
	}
	if *jsonOut {
		return errExit(render.JSON(os.Stdout, v))
	}
	return errExit(render.Pretty(os.Stdout, render.UseColor(os.Stdout), v))
}
