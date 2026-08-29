package dev.kithcraft.mod.death;

import dev.kithcraft.mod.percept.Place;
import dev.kithcraft.mod.percept.Testimony;

import java.util.Map;
import java.util.Objects;

/**
 * T009 (FR-005, card AC #6): the optional "tend &lt;name&gt;'s grave" posting. Its content is
 * §4.7's {@code text} shape ({@link Testimony#textContent}) — "something read from an
 * artifact" — regardless of what it hangs on.
 *
 * <p><b>TASK-0020 T006 closed the seam this class's own doc used to flag:</b> V4's board
 * ({@code dev.kithcraft.mod.board.Board}) is merged, and {@code
 * dev.kithcraft.mod.live.LiveDeathHandling} now posts this entry's {@link #content()} text
 * onto the real board via {@code Board#postNotice} — no longer just composed and logged. This
 * class itself needed no rework: the seam was always the text, never the fixture it hung on.
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
