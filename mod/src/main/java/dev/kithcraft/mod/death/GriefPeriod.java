package dev.kithcraft.mod.death;

import java.util.Objects;

/**
 * T008 (FR-006, card AC #7): the grief-period hold — death-mechanics.md §2's "leave the bed and
 * job-site unclaimed for a grief period" ruling, confirmed by
 * {@code specs/019-death-remains/research/death-26.2.md} R-4 to need an EXPLICIT vendor-side
 * hold (26.2 has no natural POI re-claim cooldown at all; a freed bed is re-claimable within
 * ~1-2 real seconds of {@code die()} returning). Pure core (no Minecraft types): the config
 * value and the hold-expiry arithmetic are unit-testable without a live world;
 * {@code dev.kithcraft.mod.live.LiveDeathHandling} is what actually re-claims and later
 * releases the real {@code PoiManager} ticket.
 */
public final class GriefPeriod {
    private GriefPeriod() {}

    /** Default: one day/night cycle (R-3), ~24000 ticks. A config value, not a constant baked
     * into call sites (card AC #7) — read from a system property once per call so a dev-server
     * run can retune it without a recompile, the same idiom {@code BodySession} uses for its
     * socket path. */
    public static long configuredTicks() {
        return Long.getLong("kithcraft.griefPeriodTicks", 24000L);
    }

    /** One held POI: {@code referent} is a human-readable descriptor for logging/diagnostics
     * only, never a wire field. Held while {@code currentTick < heldUntilTick}. */
    public record Hold(String referent, long heldUntilTick) {
        public Hold {
            Objects.requireNonNull(referent, "referent");
        }

        public boolean isHeld(long currentTick) {
            return currentTick < heldUntilTick;
        }
    }

    public static Hold start(String referent, long startTick) {
        return new Hold(referent, startTick + configuredTicks());
    }
}
