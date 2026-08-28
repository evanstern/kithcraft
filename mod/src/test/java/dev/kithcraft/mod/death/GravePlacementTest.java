package dev.kithcraft.mod.death;

import org.junit.jupiter.api.Test;

import java.util.Set;

import static org.junit.jupiter.api.Assertions.assertEquals;

class GravePlacementTest {

    @Test
    void deathSiteItselfWhenSafe() {
        var death = new GravePlacement.Pos(0, 64, 0);
        var found = GravePlacement.findGraveSite(death, p -> true);
        assertEquals(death, found);
    }

    @Test
    void nearestSafeRingWhenDeathSiteUnsafe() {
        // Death over "lava" (unsafe); one safe tile two rings out.
        var death = new GravePlacement.Pos(0, 64, 0);
        var safeTarget = new GravePlacement.Pos(2, 64, 0);
        var found = GravePlacement.findGraveSite(death, p -> p.equals(safeTarget));
        assertEquals(safeTarget, found);
    }

    @Test
    void deterministicTieBreakPicksSmallestRingThenFixedScanOrder() {
        // Two equally-near candidates on the same ring; the fixed dx/dz scan order always
        // picks the same one given the same safe set (no randomness, no host-dependent order).
        var death = new GravePlacement.Pos(0, 64, 0);
        var candidates = Set.of(new GravePlacement.Pos(-1, 64, 1), new GravePlacement.Pos(1, 64, -1));
        var first = GravePlacement.findGraveSite(death, candidates::contains);
        var second = GravePlacement.findGraveSite(death, candidates::contains);
        assertEquals(first, second, "search order must be deterministic across calls");
        assertEquals(new GravePlacement.Pos(-1, 64, 1), first, "fixed dx-then-dz ascending scan");
    }

    @Test
    void fallsBackToDeathSiteWhenNothingSafeInRange() {
        var death = new GravePlacement.Pos(0, 64, 0);
        var found = GravePlacement.findGraveSite(death, p -> false, 2);
        assertEquals(death, found, "card AC #4: the grave is always placed, even unsafely");
    }

    @Test
    void boundedSearchNeverExceedsMaxRadius() {
        var death = new GravePlacement.Pos(0, 64, 0);
        var farAway = new GravePlacement.Pos(50, 64, 0);
        var found = GravePlacement.findGraveSite(death, p -> p.equals(farAway), 3);
        assertEquals(death, found, "a safe tile outside maxRadius must not be found");
    }
}
