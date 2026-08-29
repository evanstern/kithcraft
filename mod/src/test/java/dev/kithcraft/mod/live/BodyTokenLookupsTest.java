package dev.kithcraft.mod.live;

import org.junit.jupiter.api.Test;

import java.util.Optional;
import java.util.UUID;
import java.util.function.Function;

import static org.junit.jupiter.api.Assertions.assertEquals;

/** TASK-0021 T003: {@link BodyTokenLookups#combine} is the fix for the token-namespace
 * mismatch {@code specs/020-job-board/research/board-observation.md} §4 found — proven here as
 * pure {@code UUID}/{@code Optional} plumbing, independent of DuskPairing/BodySession/Minecraft
 * entirely. */
class BodyTokenLookupsTest {

    private static final UUID KNOWN_TO_PRIMARY = UUID.randomUUID();
    private static final UUID KNOWN_TO_SECONDARY = UUID.randomUUID();
    private static final UUID KNOWN_TO_NEITHER = UUID.randomUUID();

    private static final Function<UUID, Optional<String>> PRIMARY =
        id -> id.equals(KNOWN_TO_PRIMARY) ? Optional.of("seat-token") : Optional.empty();
    private static final Function<UUID, Optional<String>> SECONDARY =
        id -> id.equals(KNOWN_TO_SECONDARY) ? Optional.of("live-attach-token") : Optional.empty();

    @Test
    void primaryAnswerWinsWhenBothCouldAnswer() {
        Function<UUID, Optional<String>> both =
            id -> id.equals(KNOWN_TO_PRIMARY) ? Optional.of("primary") : Optional.empty();
        Function<UUID, Optional<String>> combined = BodyTokenLookups.combine(both, uuid -> Optional.of("secondary"));
        assertEquals(Optional.of("primary"), combined.apply(KNOWN_TO_PRIMARY));
    }

    @Test
    void fallsBackToSecondaryWhenPrimaryHasNoAnswer() {
        Function<UUID, Optional<String>> combined = BodyTokenLookups.combine(PRIMARY, SECONDARY);
        assertEquals(Optional.of("live-attach-token"), combined.apply(KNOWN_TO_SECONDARY));
    }

    @Test
    void emptyWhenNeitherNamespaceKnowsTheUUID() {
        Function<UUID, Optional<String>> combined = BodyTokenLookups.combine(PRIMARY, SECONDARY);
        assertEquals(Optional.empty(), combined.apply(KNOWN_TO_NEITHER));
    }

    /** The exact bug named in board-observation.md §4: a live claim's body token (the
     * secondary/live-attach namespace here) must be found even though the seat-token namespace
     * (DuskPairing's own) has never heard of it. */
    @Test
    void findsALiveClaimantTokenTheSeatNamespaceNeverRegistered() {
        UUID liveOnlyVillager = UUID.randomUUID();
        Function<UUID, Optional<String>> seatTokensOnly = id -> Optional.empty(); // DuskPairing never seated this UUID
        Function<UUID, Optional<String>> liveAttach =
            id -> id.equals(liveOnlyVillager) ? Optional.of("b-live-claim-token") : Optional.empty();

        Function<UUID, Optional<String>> combined = BodyTokenLookups.combine(seatTokensOnly, liveAttach);
        assertEquals(Optional.of("b-live-claim-token"), combined.apply(liveOnlyVillager));
    }
}
