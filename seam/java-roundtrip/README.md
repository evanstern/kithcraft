# java-roundtrip

The JVM half of the seam wire's executable pinning.

```sh
cd seam/java-roundtrip
java RoundTrip.java ../vectors
```

Single file, launched in source-file mode — no Gradle, no Maven, no build step, no dependency.
JDK ≥ 17 for the language features used; verified on JDK 26. The vector directory argument
defaults to `../vectors`.

Exit code is 0 when every check passes and 1 otherwise, with each failure printed. The last line
is always the summary:

```
java-roundtrip: 91 passed, 0 failed, over 17 vectors
```

**Why the JSON handling is hand-rolled.** The canonical form (`docs/design/seam-wire-v0.md`
§2.4) is a small closed subset — objects, arrays, strings, integers, three literals — and a
library would mean a build system this task must not introduce. It would also weaken the point:
the harness exists to be an *independent* implementation, and two harnesses leaning on the same
well-known JSON library would be agreeing about that library rather than about the spec. It
shares no code with `../go-roundtrip` by design.

This is throwaway-grade proof, not the transport. V1 (TASK-0009) builds the real vendor side; it
should read these vectors and delete nothing.
