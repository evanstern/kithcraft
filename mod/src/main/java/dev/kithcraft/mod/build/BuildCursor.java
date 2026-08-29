package dev.kithcraft.mod.build;

import java.util.Objects;

/**
 * T009 (spec.md US4, card AC #6): the persisted resume point for a claimed build — V1's
 * SavedData-snapshot idiom ({@code Board.Snapshot}, {@code TokenRegistry.Snapshot}) applied to
 * build progress. {@code blueprintText} is captured once, at {@link #start}, not re-read from
 * the board every tick — spec Edge Cases' "board book removed/edited mid-build: the claimed
 * build continues from its captured blueprint snapshot; new reads see the new content."
 *
 * <p>{@code placed} is the only thing an interrupt can leave stale, and {@link BuildEngine}
 * treats both interrupt kinds (a dusk schedule transition, or a danger-driven PANIC) as the
 * exact same "stopped being {@code Activity.WORK}-active" signal — a later work period's
 * {@link BuildEngine#advance} resumes at the same {@code placed} index against the same
 * captured text, with no re-claim and no player action.
 */
public record BuildCursor(String claimantBody, String blueprintText, int placed) {
    public BuildCursor {
        Objects.requireNonNull(claimantBody, "claimantBody");
        Objects.requireNonNull(blueprintText, "blueprintText");
    }

    public static BuildCursor start(String claimantBody, String blueprintText) {
        return new BuildCursor(claimantBody, blueprintText, 0);
    }
}
