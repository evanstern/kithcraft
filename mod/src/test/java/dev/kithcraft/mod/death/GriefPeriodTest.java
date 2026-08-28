package dev.kithcraft.mod.death;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

class GriefPeriodTest {

    @AfterEach
    void clearOverride() {
        System.clearProperty("kithcraft.griefPeriodTicks");
    }

    @Test
    void defaultsToOneDayNightCycle() {
        assertEquals(24000L, GriefPeriod.configuredTicks(), "R-3: default one day/night cycle");
    }

    @Test
    void configOverridableNotAConstant() {
        System.setProperty("kithcraft.griefPeriodTicks", "100");
        assertEquals(100L, GriefPeriod.configuredTicks(), "card AC #7: config value, not a constant");
    }

    @Test
    void heldBeforeExpiryUnheldAfter() {
        var hold = GriefPeriod.start("Tam's bed", 1000L);
        assertEquals(1000L + 24000L, hold.heldUntilTick());
        assertTrue(hold.isHeld(1000L));
        assertTrue(hold.isHeld(1000L + 24000L - 1));
        assertFalse(hold.isHeld(1000L + 24000L));
        assertFalse(hold.isHeld(1000L + 30000L));
    }
}
