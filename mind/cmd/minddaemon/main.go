// Command minddaemon is the mind daemon's process entrypoint. This is the M1
// skeleton stub: it parses the socket-path flag and exits cleanly. The UDS
// listener and session lifecycle are Phase 2 (T006-T008).
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	socket := flag.String("socket", "", "path to the UDS socket the daemon listens on")
	flag.Parse()

	if *socket == "" {
		fmt.Fprintln(os.Stderr, "minddaemon: --socket is required")
		os.Exit(2)
	}

	fmt.Printf("minddaemon: stub build, socket=%s (listener not yet implemented)\n", *socket)
}
