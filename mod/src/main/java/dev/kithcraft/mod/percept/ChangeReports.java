package dev.kithcraft.mod.percept;

import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.Locale;
import java.util.Map;
import java.util.Objects;
import java.util.Set;

/**
 * §4.10 {@code change_report}: something changed at a place, reported to a body that was not
 * there. {@link #recipients} IS the delivery restriction, normative and load-bearing:
 *
 * <blockquote>A vendor MUST NOT deliver a {@code change_report} to a body that caused the
 * change, or that was in range to witness it. Those bodies were already told: the actor by its
 * {@code act_result} (§5.4), the witnesses by a {@code sighting} at the moment it happened.
 * </blockquote>
 *
 * <p>promptworld I shipped the naive version first and measured, live, that <b>75% of all
 * memories formed came from that channel</b> — bookkeeping noise (the world "correcting" its
 * own stale map) voiced as experience, including to the actor who caused the change and every
 * bystander who watched it. This method is the fix: it is the only way a caller composes a
 * recipient list for this percept type, so the restriction cannot be bypassed by mailing the
 * content directly.
 */
public final class ChangeReports {
    private ChangeReports() {}

    /** §4.10's closed {@code change} vocabulary. */
    public enum Change {
        APPEARED, GONE, ALTERED;

        public String wire() {
            return name().toLowerCase(Locale.ROOT);
        }
    }

    public static Map<String, Object> content(Change change, Thing thing) {
        Objects.requireNonNull(change, "change");
        Objects.requireNonNull(thing, "thing");
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("change", change.wire());
        m.put("thing", thing.toPayload());
        return m;
    }

    /** §4.10's provenance, per the protocol's own example: always {@code saw}, direct, source
     * {@code null} — the vendor is reporting its own detected world diff, not relaying another
     * body's account, so it is never {@code told}. */
    public static Provenance provenance(Long observedAt, long receivedAt) {
        return Provenance.direct(Provenance.Origin.SAW, observedAt, receivedAt);
    }

    /**
     * The delivery restriction: {@code candidates} minus the {@code actor} (who caused the
     * change) minus every {@code witness} (who was in range to see it happen).
     *
     * @return the bodies that MAY receive this change_report — everyone else, per §4.10.
     */
    public static Set<String> recipients(Set<String> candidates, String actor, Set<String> witnesses) {
        Objects.requireNonNull(candidates, "candidates");
        Set<String> out = new LinkedHashSet<>(candidates);
        if (actor != null) {
            out.remove(actor);
        }
        if (witnesses != null) {
            out.removeAll(witnesses);
        }
        return out;
    }
}
