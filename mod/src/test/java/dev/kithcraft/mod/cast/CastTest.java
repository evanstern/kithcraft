package dev.kithcraft.mod.cast;

import org.junit.jupiter.api.Test;

import java.util.HashSet;
import java.util.Set;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * T002: seeding idempotence and identity survival across a restart (the unit half; the
 * dev-server restart observation is a later phase's live proof, as {@code TokenRegistryTest}
 * does for {@code TokenRegistry}).
 */
class CastTest {

    @Test
    void threeDistinctProfessionByBiomeVariantPairings() {
        assertEquals(3, Cast.MEMBERS.size());
        Set<String> pairings = new HashSet<>();
        for (Cast.Member member : Cast.MEMBERS) {
            pairings.add(member.profession() + "/" + member.biomeVariant());
        }
        assertEquals(3, pairings.size(), "each cast member must be a distinct profession x biome variant pairing");
    }

    @Test
    void freshCastIsNotSeeded() {
        assertFalse(new Cast().isSeeded());
    }

    @Test
    void markingSeededTwiceIsIdempotent() {
        Cast cast = new Cast();
        cast.markSeeded();
        cast.markSeeded();
        assertTrue(cast.isSeeded());
    }

    @Test
    void snapshotRoundTripPreservesSeededState() {
        Cast original = new Cast();
        original.markSeeded();

        Cast restored = Cast.restore(original.snapshot());

        assertTrue(restored.isSeeded(), "seeded state must survive the round trip (identity survives a restart)");
    }

    @Test
    void unseededSnapshotRoundTripStaysUnseeded() {
        Cast restored = Cast.restore(new Cast().snapshot());
        assertFalse(restored.isSeeded());
    }
}
