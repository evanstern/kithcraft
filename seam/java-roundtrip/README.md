# java-roundtrip

The JVM half of the seam wire's executable pinning.

```sh
cd mod && ./gradlew seamRoundTrip
```

```
> Task :seamRoundTrip

java-roundtrip: 91 passed, 0 failed, over 17 vectors
```

**Shape: a Gradle task in `mod/`, not a standalone build or a jar-fetching script.** T010
(the operator's 2026-08-22 ruling) required rebuilding this harness on a library once
TASK-0009 introduced Gradle — the hand-roll's only justification (no build system exists
yet) went away the moment `mod/` landed. `RoundTrip.java` stays a single file at this path,
launched via JEP 330 source-launch mode (`java <file>.java <args>`, no compile step); the
`seamRoundTrip` task in `mod/build.gradle` supplies Gson on the classpath from
`sourceSets.main.runtimeClasspath` (Gson arrives transitively through the `minecraft`
configuration — the same free ride `mod/wire/CanonicalJson.java` takes, no new dependency
anywhere) and invokes `java --class-path <that classpath> RoundTrip.java seam/vectors`.
A second standalone Gradle project here would duplicate a wrapper (binary jar) this repo
already commits once; a jar-fetching shell script would hand-roll Gradle's own dependency
resolution. Reusing the mod's already-resolved classpath is the smallest true diff.

Manual invocation (needs Gson on the classpath, e.g. any jar containing
`com.google.gson.stream.JsonReader`):

```sh
java --class-path <path-to-gson.jar> seam/java-roundtrip/RoundTrip.java seam/vectors
```

Exit code is 0 when every check passes and 1 otherwise, with each failure printed. The last
line is always the summary shown above.

**Why the writer stays hand-rolled.** Decode now uses Gson's `JsonReader` — the same
choice `mod/wire/CanonicalJson.java` made, and the same C-1..C-10 emit-check verdict
recorded in `specs/009-fabric-mod-skeleton/research/versions.md`: Gson cannot be made to
emit canonical form as a thin wrapper (no sorted-key mode for C-3, no integer-only numeric
mode for C-6), so encode stays a small hand-written writer, per the ruling's carve-out. This
harness still shares no code with `../go-roundtrip` — the point of two implementations
agreeing was always that the *spec*, not one codebase's parsing habits, is what the vectors
pin; letting both the vendor's own codec and this harness use the same well-understood
library (rather than each hand-rolling a parser) doesn't weaken that, since the wire's own
form (C-1..C-10) still has to be produced by hand in both places.

**Mutation-check power, demonstrated 2026-08-25** (against a `/tmp` copy of the vectors;
the tracked `seam/vectors/` was never touched — `git status` confirmed clean before and
after):

- *Corrupt a byte* — flipped one hex nibble mid-frame in `percept_sighting.json`'s
  `frame_hex`. Result: **red**, 2 failures (`percept_sighting/meaning` — the corrupted byte
  landed inside `percept_type`, decoding to `"ighting"` instead of `"sighting"` — and
  `percept_sighting/bytes` — the re-encode no longer reproduces the pinned hex at the
  corrupted offset). 89 passed, 2 failed, over 17 vectors.
- *Decoded-only drift* — changed only the human-authored `decoded.payload.place.descriptor`
  field (leaving `frame_hex` untouched). Result: **red**, exactly 1 failure
  (`percept_sighting/meaning` — the frame still decodes and re-encodes correctly since
  `frame_hex` is unchanged, but what it decodes to no longer matches the vector's declared
  `decoded` form). 90 passed, 1 failed, over 17 vectors. This is the harness's sharpest
  power: a decoder that quietly mangled a field and an encoder that mangled it back the
  same way would pass the `/bytes` check alone, which is exactly what `/meaning` exists to
  catch.

Both were reverted immediately after observing the failure; no vector file changed.

**Provenance.** JDK ≥ 17 for the language features used; verified on JDK 26.0.2. Was single
file, hand-rolled JSON, zero-dependency, `java RoundTrip.java ../vectors` — see git history
for that version and its rationale (throwaway proof, pre-Gradle repo). This is throwaway-
grade proof, not the transport: V1 (TASK-0009) builds the real vendor side (`mod/`); this
harness reads the vectors and deletes nothing.
