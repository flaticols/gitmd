// Command gitmd manages git commit trailers (Reason, Ticket, Assisted, …).
// See `gitmd help` for usage.
package main

import (
	"os"

	"github.com/flaticols/gitmd/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
