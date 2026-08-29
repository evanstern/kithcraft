package dev.kithcraft.mod.board;

import java.util.HashSet;
import java.util.Objects;
import java.util.Set;

/**
 * T002's dedup half — mirrors {@code dev.kithcraft.mod.brain.PairingSignal}'s "already fired
 * this approach" bookkeeping, simplified: the board's read is a proximity gate, not a
 * predicted-arrival one (no ETA math needed — a villager either is or isn't at the board this
 * tick). Fires exactly once per approach; leaving the board's radius clears the dedup so a
 * later, genuinely fresh visit can fire again.
 */
public final class BoardReadTracker {
    private final Set<String> readThisApproach = new HashSet<>();

    /** @return true the first tick {@code near} is true for {@code bodyKey} since it was
     * last false (a fresh arrival); false on every later tick of the same visit, and false
     * whenever {@code near} is false (which also clears the dedup, so leaving resets it). */
    public boolean shouldFireNow(String bodyKey, boolean near) {
        Objects.requireNonNull(bodyKey, "bodyKey");
        if (!near) {
            readThisApproach.remove(bodyKey);
            return false;
        }
        return readThisApproach.add(bodyKey);
    }
}
