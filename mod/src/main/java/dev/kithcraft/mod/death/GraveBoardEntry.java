package dev.kithcraft.mod.death;

import dev.kithcraft.mod.percept.Place;
import dev.kithcraft.mod.percept.Testimony;

import java.util.Map;
import java.util.Objects;

/**
 * T009 (FR-005, card AC #6): the optional "tend &lt;name&gt;'s grave" posting, per plan.md's V4
 * decoupling seam — V4 (the job-board, TASK-0020) is not merged, so this rides the read channel
 * V2 already carries instead of a not-yet-real book block: §4.7's {@code text} percept
 * ({@link Testimony#textContent}), the shape a mind already treats as "something read from an
 * artifact." When V4 merges its board book, the same content rides that channel with no rework
 * — the seam is the text, not the fixture it hangs on.
 *
 * <p>Takeable or ignorable (design-brief #6: reluctance/grumbling is character, not a bug) —
 * {@link #take()} exists so a survivor CAN claim it; nothing requires it to ever be called.
 */
public final class GraveBoardEntry {
    private final String villagerName;
    private final Place gravePlace;
    private boolean taken;

    public GraveBoardEntry(String villagerName, Place gravePlace) {
        this.villagerName = Objects.requireNonNull(villagerName, "villagerName");
        this.gravePlace = Objects.requireNonNull(gravePlace, "gravePlace");
    }

    /** §4.7 content: the posting text, attributed to no one in particular (a board notice, not
     * a claim from a named teller). */
    public Map<String, Object> content() {
        return Testimony.textContent("Tend " + villagerName + "'s grave.", null);
    }

    public Place gravePlace() {
        return gravePlace;
    }

    public boolean isTaken() {
        return taken;
    }

    /** Idempotent: taking an already-taken entry leaves it taken. */
    public void take() {
        taken = true;
    }
}
