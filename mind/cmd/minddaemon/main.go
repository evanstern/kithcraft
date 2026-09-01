// Command minddaemon is the mind daemon's process entrypoint: it listens on
// a UDS path (decision-0004), loads or generates the demo cast's personas,
// opens their durable stores, and runs the session lifecycle in mind/seam
// over each accepted connection (TASK-0021 T001).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"kithcraft/mind/seam"
)

func main() {
	socket := flag.String("socket", envOr("MINDDAEMON_SOCKET", ""), "path to the UDS socket the daemon listens on (env MINDDAEMON_SOCKET)")
	runDir := flag.String("rundir", envOr("MINDDAEMON_RUNDIR", "run"), "directory for persisted state (personas, memory logs, ledgers, archive) (env MINDDAEMON_RUNDIR)")
	// FR-007's rehearsal path (spec.md US1 scenario 2): off forces the
	// zero-call path unconditionally, regardless of whether
	// ANTHROPIC_API_KEY happens to be exported — the R-3/R-6 mod-side
	// knobs live in mod/ (system properties), not here.
	genesis := flag.Bool("genesis", envOrBool("MINDDAEMON_GENESIS", true), "run persona genesis for missing cast members; false forces the zero-call rehearsal path (env MINDDAEMON_GENESIS)")
	flag.Parse()

	if *socket == "" {
		fmt.Fprintln(os.Stderr, "minddaemon: --socket is required")
		os.Exit(2)
	}

	rt, err := NewRuntime(*runDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer rt.Close()
	defer rt.Report(os.Stdout, *runDir)

	if !*genesis {
		rt.Client, rt.Digester = nil, nil
	}
	if err := rt.LoadOrGenesisCast(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ln, err := listen(*socket)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer ln.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		ln.Close() // graceful shutdown: Close removes the socket file it created
	}()

	ing := seam.NewIngester()
	ing.Archived = rt.Archive.IsArchived
	ing.OnPercept = rt.HandlePercept
	ing.OnSessionOpen = rt.HandleSessionOpen

	fmt.Printf("minddaemon: listening on %s\n", *socket)
	serve(ln, ing)
}
