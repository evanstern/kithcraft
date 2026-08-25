// Command minddaemon is the mind daemon's process entrypoint: it listens on
// a UDS path (decision-0004) and runs the session lifecycle in mind/seam
// over each accepted connection.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"kithcraft/mind/seam"
)

func main() {
	socket := flag.String("socket", "", "path to the UDS socket the daemon listens on")
	flag.Parse()

	if *socket == "" {
		fmt.Fprintln(os.Stderr, "minddaemon: --socket is required")
		os.Exit(2)
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

	fmt.Printf("minddaemon: listening on %s\n", *socket)
	serve(ln, seam.NewIngester())
}
