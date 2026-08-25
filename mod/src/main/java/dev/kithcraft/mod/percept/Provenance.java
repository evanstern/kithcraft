package dev.kithcraft.mod.percept;

import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Objects;

/**
 * §2.6-2.7's provenance: the closed origin vocabulary, stamped at emission, never inferred
 * or defaulted. {@code source} is always the IMMEDIATE teller, never the original observer
 * — a relayed fact's {@code hops} grows but its {@code origin} never launders to direct.
 */
public final class Provenance {

    /** §2.7's closed origin vocabulary and EH-3's classifier: a pure function of the origin
     * value alone, never of text/descriptor/hops. A {@code direct} boolean deliberately does
     * NOT cross the wire (§2.7) — {@link #isDirect()} is a JVM-side query only. */
    public enum Origin {
        ACTED("acted", true), SAW("saw", true), HEARD("heard", true), FELT("felt", true),
        TOLD("told", false), READ("read", false);

        private final String wire;
        private final boolean direct;

        Origin(String wire, boolean direct) {
            this.wire = wire;
            this.direct = direct;
        }

        public String wire() {
            return wire;
        }

        /** EH-3: {@code acted}/{@code saw}/{@code heard}/{@code felt} are direct; {@code
         * told}/{@code read} are not. An unrecognized wire string classifies secondhand
         * (EH-2b) — callers decoding an arbitrary string, not this closed enum, must treat
         * a parse failure as non-direct rather than default to direct. */
        public boolean isDirect() {
            return direct;
        }
    }

    /** §2.6's {@code source}: who or what this reached the body through. {@code kind} ∈
     * {@code body} | {@code person} | {@code artifact} | {@code unknown}. */
    public record Source(String kind, String body, String thingId, String descriptor) {
        public Source {
            Objects.requireNonNull(kind, "kind");
        }

        public static Source body(String bodyToken, String descriptor) {
            return new Source("body", Objects.requireNonNull(bodyToken, "bodyToken"), null, descriptor);
        }

        public static Source person(String descriptor) {
            return new Source("person", null, null, descriptor);
        }

        public static Source artifact(String thingId, String descriptor) {
            return new Source("artifact", null, Objects.requireNonNull(thingId, "thingId"), descriptor);
        }

        public static Source unknown(String descriptor) {
            return new Source("unknown", null, null, descriptor);
        }

        Map<String, Object> toPayload() {
            Map<String, Object> m = new LinkedHashMap<>();
            m.put("kind", kind);
            if (body != null) {
                m.put("body", body);
            }
            if (thingId != null) {
                m.put("thing_id", thingId);
            }
            m.put("descriptor", descriptor);
            return m;
        }
    }

    public final Origin origin;
    public final Source source;    // null only for the four direct origins
    public final Long observedAt;  // null = unknown = maximally stale (§2.2)
    public final long receivedAt;
    public final Integer hops;     // informational only (§2.6); omitted from the wire when null

    private Provenance(Origin origin, Source source, Long observedAt, long receivedAt, Integer hops) {
        this.origin = Objects.requireNonNull(origin, "origin");
        if (origin.isDirect() && source != null) {
            throw new IllegalArgumentException(
                origin.wire() + " is a direct origin: source must be null — the body is its own witness (§2.6)");
        }
        if (!origin.isDirect() && source == null) {
            throw new IllegalArgumentException(
                origin.wire() + " is an indirect origin: source (the immediate teller) is required (§2.6)");
        }
        this.source = source;
        this.observedAt = observedAt;
        this.receivedAt = receivedAt;
        this.hops = hops;
    }

    /** A direct origin (acted/saw/heard/felt): the body is its own witness, stamped at
     * emission. {@code observedAt} MAY be {@code null} only where genuinely unknown. */
    public static Provenance direct(Origin origin, Long observedAt, long receivedAt) {
        return new Provenance(origin, null, observedAt, receivedAt, null);
    }

    /** An indirect origin (told/read): {@code source} is the IMMEDIATE teller, never the
     * original observer. {@code observedAt} for a told fact is the TELLER's last-seen time,
     * not the telling time (EH-5) — pass that value here, not {@code receivedAt}. */
    public static Provenance indirect(Origin origin, Source source, Long observedAt, long receivedAt, Integer hops) {
        Objects.requireNonNull(source, "source");
        return new Provenance(origin, source, observedAt, receivedAt, hops);
    }

    public Map<String, Object> toPayload() {
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("origin", origin.wire());
        m.put("source", source == null ? null : source.toPayload());
        m.put("observed_at", observedAt);
        m.put("received_at", receivedAt);
        if (hops != null) {
            m.put("hops", hops.longValue());
        }
        return m;
    }

    @Override
    public String toString() {
        return "Provenance" + toPayload();
    }
}
