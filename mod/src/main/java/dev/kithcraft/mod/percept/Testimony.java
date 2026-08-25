package dev.kithcraft.mod.percept;

import java.util.LinkedHashMap;
import java.util.Locale;
import java.util.Map;
import java.util.Objects;

/**
 * §4.6 {@code told_fact} and §4.7 {@code text}: the two indirect-knowledge channels — a fact
 * passed body-to-body, and text read from an artifact. Both are {@code indirect} provenance
 * (§2.7, {@link Provenance#indirect}); the trust semantics differ per type and are documented
 * at each factory.
 */
public final class Testimony {
    private Testimony() {}

    /** §4.6's closed {@code assertion} vocabulary. */
    public enum Assertion {
        PRESENT, GONE;

        public String wire() {
            return name().toLowerCase(Locale.ROOT);
        }
    }

    /** §4.6's content: what the teller asserted, about a place and/or a thing. */
    public static Map<String, Object> toldFactContent(Place aboutPlace, Thing thing, Assertion assertion) {
        Objects.requireNonNull(assertion, "assertion");
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("about_place", aboutPlace == null ? null : aboutPlace.toPayload());
        m.put("thing", thing == null ? null : thing.toPayload());
        m.put("assertion", assertion.wire());
        return m;
    }

    /** §4.6's provenance: {@code told}, indirect, {@code source} the IMMEDIATE teller (never
     * the original observer — {@code hops} grows, {@code origin} stays {@code told}).
     * {@code tellerLastSeenAt} is the TELLER's last-seen time (EH-5), not the telling time —
     * pass that here, never {@code receivedAt}. The anti-telephone-game comparison this
     * enables ("a receiver's own fresher knowledge never loses to secondhand") is mind-side;
     * this method's only obligation is putting the honest teller timestamp on the wire. */
    public static Provenance toldFactProvenance(
            Provenance.Source teller, Long tellerLastSeenAt, long receivedAt, Integer hops) {
        return Provenance.indirect(Provenance.Origin.TOLD, teller, tellerLastSeenAt, receivedAt, hops);
    }

    /** §4.7's content: {@code attributedTo} is what the artifact CLAIMS, not what the vendor
     * verified. A mind MUST NOT treat it as {@code told} provenance from that party. */
    public static Map<String, Object> textContent(String text, String attributedTo) {
        Objects.requireNonNull(text, "text");
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("text", text);
        m.put("attributed_to", attributedTo);
        return m;
    }

    /** §4.7's provenance: {@code read}, indirect, {@code source} the artifact.
     * {@code observedAt} is typically {@code null} — an artifact's authorship time is usually
     * unknown, meaning maximally stale (§2.2). */
    public static Provenance textProvenance(Provenance.Source artifact, Long observedAt, long receivedAt) {
        return Provenance.indirect(Provenance.Origin.READ, artifact, observedAt, receivedAt, null);
    }
}
