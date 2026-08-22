# Quickstart — validating the seam wire pinning

Prerequisites: Go toolchain; JDK ≥ 17. No network, no Minecraft, no API key.

Both harnesses read the same seventeen fixtures in `seam/vectors/` and share no code. Two
independent implementations agreeing on every byte is the evidence; either one alone would only
be agreeing with itself.

## Run the Go round-trip

```sh
cd seam/go-roundtrip
go test ./...
```

Observed (go1.26.4, darwin/arm64):

```
ok  	kithcraft/seam/go-roundtrip	0.331s
```

With `-v`, the census line and 17 named subtests:

```
--- PASS: TestCensus (0.00s)
    roundtrip_test.go:90: census: 17 vectors, matching contracts/vectors.md exactly
--- PASS: TestRoundTrip (0.00s)
    --- PASS: TestRoundTrip/cancel (0.00s)
    --- PASS: TestRoundTrip/err_missing_provenance (0.00s)
    … 15 more, one per vector …
--- PASS: TestFramingRefusals (0.00s)
--- PASS: TestReceiverAcceptsNonCanonical (0.00s)
```

Each `TestRoundTrip` subtest is the full §6 obligation for one vector: decode the pinned frame,
match the decoded form structurally, re-encode canonically, compare bytes over the whole frame
including the four-byte length word, then check the vector's pinned validation behavior.

## Run the Java round-trip

```sh
cd seam/java-roundtrip
java RoundTrip.java ../vectors
```

Single file, source-file mode — no Gradle, no build step. Observed (OpenJDK 26.0.2):

```
java-roundtrip: 91 passed, 0 failed, over 17 vectors
```

The 91 is the check count, not the vector count: the census contributes 17, each vector
contributes four (decode, meaning, bytes, validation behavior), and six more cover the
frame-layer refusals and the sender/receiver asymmetry. Exit code is 0 on green, 1 on any
failure, with each failure printed above the summary.

## Confirming the harnesses can fail

Both were mutation-checked while being written, since a suite that has never gone red is a suite
with no demonstrated power to catch anything:

- **One hex byte corrupted** in `percept_speech.json` → Go `--- FAIL: TestRoundTrip/percept_speech`,
  Java `87 passed, 1 failed`. Both reported the byte offset of the first difference.
- **`hops: 0` changed to `hops: 3` in the decoded form only**, leaving the bytes untouched → Go
  `decoded value differs from the vector's declared form`, Java
  `FAIL percept_speech/meaning: decoded value differs from the declared form`.

The second is the one worth having: it is caught only by the structural assertion, and it is
exactly the failure a byte comparison alone would have waved through.

## Audit the deliverables

- Decision record exists under `backlog/decisions/` with the T-matrix filled
  (T-1..T-7 × {UDS, TCP, stdio}), status proposed.
- `docs/design/seam-wire-v0.md` exists with all seven sections per data-model.md.
- `seam/vectors/` contains exactly the vector set contracts/vectors.md closes over — both
  harnesses enforce this in both directions, so a missing *or* extra vector fails the run.
- Leak check: no engine-native type/identifier/coordinate convention in the framing spec or any
  vector (grep for obvious offenders; §12 method).
- Diff check: nothing outside the five artifact groups (decision, framing spec, vectors, two
  harnesses, spec-dir paper trail).
