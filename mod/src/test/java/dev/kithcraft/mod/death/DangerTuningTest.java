package dev.kithcraft.mod.death;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * T004 (card ACs #4/#6): the R-6 knob is config, off by default, and a recorded-take-only
 * opt-in — not a code constant. Mirrors {@link GriefPeriodTest}'s system-property idiom.
 */
class DangerTuningTest {

    @AfterEach
    void clearOverride() {
        System.clearProperty("kithcraft.dangerTuning");
    }

    @Test
    void offByDefault() {
        assertFalse(DangerTuning.enabled(), "R-6: danger tuning is a per-run choice, never the default");
    }

    @Test
    void configOverridableNotAConstant() {
        System.setProperty("kithcraft.dangerTuning", "true");
        assertTrue(DangerTuning.enabled(), "card AC #6: config value, not a constant");
    }

    @Test
    void suppressesOnlyWhenEnabledAndWithinRadius() {
        assertFalse(DangerTuning.shouldSuppress(false, 1.0), "disabled: never suppresses, however close");
        assertFalse(DangerTuning.shouldSuppress(true, DangerTuning.SUPPRESSION_RADIUS + 1), "enabled but out of radius: not suppressed");
        assertTrue(DangerTuning.shouldSuppress(true, DangerTuning.SUPPRESSION_RADIUS), "enabled and at the radius edge: suppressed");
        assertFalse(DangerTuning.shouldSuppress(true, Double.POSITIVE_INFINITY), "no cast member loaded: never suppressed");
    }
}
