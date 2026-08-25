package dev.kithcraft.mod.percept;

import java.util.LinkedHashMap;
import java.util.Locale;
import java.util.Map;
import java.util.Objects;

/** §4.9 {@code self_state}: the body's own condition. Always {@code felt}-origin, direct —
 * the body is the only possible witness to its own condition (§2.7). */
public final class SelfState {
    private SelfState() {}

    // §4.9's open-vocabulary seeds. Minds MUST tolerate unknown conditions (an unrecognized
    // string is not an error), so these are suggested labels, not a closed enum.
    public static final String HUNGER = "hunger";
    public static final String INJURY = "injury";
    public static final String COLD = "cold";
    public static final String HEAT = "heat";
    public static final String FATIGUE = "fatigue";
    public static final String ENCUMBERED = "encumbered";
    public static final String THREATENED = "threatened";

    /** AR-6's condition band — never a raw stat. */
    public enum Level { NONE, MILD, SEVERE, CRITICAL }

    public enum Trend { WORSENING, STEADY, EASING }

    public static Map<String, Object> content(String condition, Level level, Trend trend) {
        Objects.requireNonNull(condition, "condition");
        Objects.requireNonNull(level, "level");
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("condition", condition);
        m.put("level", level.name().toLowerCase(Locale.ROOT));
        m.put("trend", trend == null ? null : trend.name().toLowerCase(Locale.ROOT));
        return m;
    }

    /** §4.9's provenance is always {@code felt}, direct — {@code observedAt} MAY be {@code
     * null} only where genuinely unknown (§2.2). */
    public static Provenance provenance(Long observedAt, long receivedAt) {
        return Provenance.direct(Provenance.Origin.FELT, observedAt, receivedAt);
    }
}
