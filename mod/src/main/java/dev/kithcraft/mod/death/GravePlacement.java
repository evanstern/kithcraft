package dev.kithcraft.mod.death;

import java.util.Objects;

/**
 * T007 (FR-004, card AC #4): the death-site placement search — death-mechanics.md §2's "at the
 * death location, defaulting to the nearest safe, buildable surface" edge case (over lava/void).
 * Pure core (plain int triples, no Minecraft types) so the deterministic search order is
 * unit-testable without a live world; {@code dev.kithcraft.mod.live.LiveDeathHandling} supplies
 * the real safety predicate (solid, non-liquid ground; clear, non-liquid space above it) and
 * turns the result back into a {@code BlockPos}.
 *
 * <p>ponytail: the search prefers the death column, then rings outward on the horizontal plane,
 * dipping at most two blocks below the death Y before widening the ring — enough to clear a
 * lava lake's shoreline or a void gap within a stated bound, not a full 3D flood-fill. Ceiling:
 * correct for the "hovering over lava/void near solid ground" edge case card AC #4 names; a
 * cliff-overhang or fully enclosed void would need a real pathfinding-grade search, which this
 * mod's scope (a permadeath demo, not a construction AI) has never asked for.
 */
public final class GravePlacement {
    private GravePlacement() {}

    /** A plain block coordinate — kept dependency-free so this class needs no Minecraft type. */
    public record Pos(int x, int y, int z) {
        public Pos offset(int dx, int dy, int dz) {
            return new Pos(x + dx, y + dy, z + dz);
        }
    }

    @FunctionalInterface
    public interface SafeCheck {
        boolean isSafeBuildable(Pos pos);
    }

    /** The default bounded search radius (blocks), matching the "nearest" clause's intent —
     * far enough to clear a shoreline, not an unbounded world scan. */
    public static final int DEFAULT_MAX_RADIUS = 8;

    /**
     * Deterministic, bounded search for a grave site: the death position itself first; then
     * rings of increasing horizontal (Chebyshev) distance, each ring checked at the death Y,
     * then one and two blocks below it, in that fixed order. Falls back to the death position
     * itself if nothing safe is found within {@code maxRadius} — the grave is placed either
     * way (card AC #4: "always"); an unsafe fallback site is the documented deviation for the
     * edge case where no safe surface exists in range at all.
     */
    public static Pos findGraveSite(Pos death, SafeCheck safe, int maxRadius) {
        Objects.requireNonNull(death, "death");
        Objects.requireNonNull(safe, "safe");
        if (safe.isSafeBuildable(death)) {
            return death;
        }
        for (int r = 1; r <= maxRadius; r++) {
            for (int dy = 0; dy >= -2; dy--) {
                for (int dx = -r; dx <= r; dx++) {
                    for (int dz = -r; dz <= r; dz++) {
                        if (Math.max(Math.abs(dx), Math.abs(dz)) != r) {
                            continue; // only this ring's boundary — inner cells were checked at a smaller r
                        }
                        Pos candidate = death.offset(dx, dy, dz);
                        if (safe.isSafeBuildable(candidate)) {
                            return candidate;
                        }
                    }
                }
            }
        }
        return death;
    }

    public static Pos findGraveSite(Pos death, SafeCheck safe) {
        return findGraveSite(death, safe, DEFAULT_MAX_RADIUS);
    }
}
