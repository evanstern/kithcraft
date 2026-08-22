# go-roundtrip

The Go half of the seam wire's executable pinning.

```sh
cd seam/go-roundtrip
go test ./...
```

Own minimal `go.mod`, stdlib only, no dependency. Verified on go1.26.

Three test functions: `TestCensus` (the vector set matches `contracts/vectors.md` exactly, in
both directions), `TestRoundTrip` (decode → structural match → canonical re-encode → byte
equality over the whole frame, plus each vector's pinned validation behavior), and
`TestFramingRefusals` / `TestReceiverAcceptsNonCanonical` (the frame-layer and asymmetry rules no
fixture can carry as data, because a fixture for a malformed frame would be a fixture that is not
a frame).

**Authoring mode.** `go test ./... -update` regenerates each vector's `frame_hex` from its
decoded form, rewriting only that line. Never run it to make a failure go away — see
`../vectors/README.md`.

**Why the canonical encoder is hand-rolled** rather than `encoding/json`: the stdlib encoder
violates C-7 twice, escaping `<`, `>`, `&`, U+2028, and U+2029 that RFC 8259 does not require
escaped, and spelling `\b` and `\f` as `\u00xx`. Both would produce bytes no conforming peer
would reproduce.

This is throwaway-grade proof, not the transport. M1 (TASK-0008) builds the real mind side.
