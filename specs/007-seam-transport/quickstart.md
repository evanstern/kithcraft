# Quickstart — validating the seam wire pinning

Prerequisites: Go toolchain; JDK ≥ 17. No network, no Minecraft, no API key.

## Run the Go round-trip

```sh
cd seam/go-roundtrip
go test ./...
```

Expected: every vector in `seam/vectors/` round-trips per contracts/vectors.md; zero
failures.

## Run the Java round-trip

```sh
cd seam/java-roundtrip
# exact invocation documented in its README once written; single-file, plain `java`
java RoundTrip.java ../vectors
```

Expected: same — every vector decodes and re-encodes to equality; error vectors refuse
as pinned.

## Audit the deliverables

- Decision record exists under `backlog/decisions/` with the T-matrix filled
  (T-1..T-7 × {UDS, TCP, stdio}), status proposed.
- `docs/design/seam-wire-v0.md` exists with all seven sections per data-model.md.
- `seam/vectors/` contains exactly the vector set contracts/vectors.md closes over.
- Leak check: no engine-native type/identifier/coordinate convention in the framing
  spec or any vector (grep for obvious offenders; §12 method).
- Diff check: nothing outside the five artifact groups (decision, framing spec,
  vectors, two harnesses, spec-dir paper trail).
