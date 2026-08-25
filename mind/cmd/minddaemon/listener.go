package main

import (
	"fmt"
	"net"
	"os"

	"kithcraft/mind/seam"
)

// maxSocketPathBytes is sun_path's usable length (104-byte field, last byte
// the terminator; decision-0004, seam-wire-v0.md §1.1). Both sides must
// fail loudly on a longer path rather than truncate it.
const maxSocketPathBytes = 103

// listen implements decision-0004's bind sequence: dial the existing path
// first (the liveness probe, W-7) and unlink only when that dial fails — a
// successful dial means another daemon holds the path, which is a startup
// error, never something to steal.
func listen(path string) (net.Listener, error) {
	if len(path) > maxSocketPathBytes {
		return nil, fmt.Errorf("minddaemon: socket path %d bytes exceeds the %d-byte sun_path limit", len(path), maxSocketPathBytes)
	}
	if conn, err := net.Dial("unix", path); err == nil {
		conn.Close()
		return nil, fmt.Errorf("minddaemon: socket %s is already live (another daemon holds it)", path)
	}
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return nil, fmt.Errorf("minddaemon: removing stale socket %s: %w", path, err)
		}
	}
	return net.Listen("unix", path)
}

// serve accepts connections for the daemon's whole life (seam-wire-v0.md
// §1.2): a vendor disconnect is normal, never terminal, and the loop only
// ends when ln itself closes (graceful shutdown).
func serve(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go func() {
			if err := seam.HandleConnection(seam.NewWireConn(conn)); err != nil {
				fmt.Fprintf(os.Stderr, "minddaemon: connection ended: %v\n", err)
			}
		}()
	}
}
