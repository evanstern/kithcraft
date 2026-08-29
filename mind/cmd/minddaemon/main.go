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
	socket := flag.String("socket", "", "path to the UDS socket the daemon listens on")
	// ponytail: this is the whole config surface Phase 1 needs to run for
	// real; env/flag knobs for R-3/R-6 and a genesis on/off switch are
	// TASK-0021 T004 (Phase 2)'s job, not duplicated here.
	runDir := flag.String("rundir", "run", "directory for persisted state (personas, memory logs, ledgers, archive)")
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

	fmt.Printf("minddaemon: listening on %s\n", *socket)
	serve(ln, ing)
}
