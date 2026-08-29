package dev.kithcraft.mod.live;

import java.util.Optional;
import java.util.UUID;
import java.util.function.Function;

/**
 * TASK-0021 T003: closes the token-namespace mismatch
 * {@code specs/020-job-board/research/board-observation.md} §4 found — {@code
 * LiveBuildExecution#findClaimant} (and {@code BoardVisit#tick} beside it) resolve a villager's
 * body token through a single {@code Function<UUID, Optional<String>>}, but the only lookup
 * ever wired in was {@code DuskPairing#bodyTokenFor} — the dusk-gathering SEAT token namespace.
 * A live claim, though, rides {@link BodySession}'s own, separate {@code ground.issueBody}
 * token. Two disjoint namespaces on either side of one {@code equals} check meant {@code
 * board.claimedBy()} could never match any villager a live claim actually came from.
 *
 * <p>The fix is exactly what the observation named as the glue-sized option: consult both
 * namespaces. {@link #combine} tries the seat-token lookup first (the common case once more
 * than one body is live-attached at once), falling back to the single live-attached body's own
 * token — pure {@code UUID}/{@code Optional} plumbing, no Minecraft type involved, so it is
 * unit-testable without a game bootstrap.
 */
public final class BodyTokenLookups {
    private BodyTokenLookups() {
    }

    /** Returns a lookup that tries {@code primary} first and falls back to {@code secondary}
     * only when {@code primary} has no answer for that UUID. Neither side is consulted beyond
     * what it already reports — this composes two closed namespaces, it does not merge them. */
    public static Function<UUID, Optional<String>> combine(
            Function<UUID, Optional<String>> primary, Function<UUID, Optional<String>> secondary) {
        return uuid -> {
            Optional<String> found = primary.apply(uuid);
            return found.isPresent() ? found : secondary.apply(uuid);
        };
    }
}
