package dev.kithcraft.mod.board;

import dev.kithcraft.mod.percept.Testimony;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Objects;

/**
 * T001 (FR-001, card ACs #1, #7): the job board's posting — free text, no form, no syntax.
 * The board is a designated lectern at a fixed config position ({@link BoardSetup} — the
 * simpler of plan.md's two options; no block-place event hook or Mixin needed to detect "the
 * first lectern placed"). Whatever the player writes into the lectern's book IS the blueprint;
 * this class never parses it (Phase 3's fidelity-floor parser reads this same text later —
 * Phase 1 only carries it whole).
 *
 * <p>T003 (card AC #4's read half): claims made on a posting are visible to every villager
 * that reads the board — modeled now as a claims section appended to the readable content.
 * Phase 2 is what actually calls {@link #recordClaim}; Phase 1 only builds the surface so the
 * read cycle already carries claims once they exist.
 */
public final class Board {
    private String text;
    private final List<String> claims = new ArrayList<>();

    /** Snapshot for {@link BoardData}'s persistence — a plain value, no Minecraft types,
     * mirroring {@code TokenRegistry.Snapshot}'s own shape. */
    public record Snapshot(String text, List<String> claims) {
        public Snapshot {
            claims = List.copyOf(claims);
        }
    }

    /** Replaces the current posting. {@code null} means no posting has ever been made —
     * {@link #readableContent()} then carries empty text, still a valid (if uninformative)
     * read. */
    public void post(String text) {
        this.text = text;
    }

    public String text() {
        return text;
    }

    public List<String> claims() {
        return List.copyOf(claims);
    }

    /** Phase 2's hook: appends one claim line to the board's visible content. This class only
     * carries lines — it does not dedupe or validate them (that's Phase 2's job once real
     * claim identity exists). */
    public void recordClaim(String claimLine) {
        claims.add(Objects.requireNonNull(claimLine, "claimLine"));
    }

    /** §4.7's {@code text} content: the posting, plus any claims appended as board-notice
     * lines — one artifact, one read, the same channel the blueprint itself rides (Q-6).
     * Attributed to "the player" per the protocol's own worked example (§4.7) — the board's
     * posting mechanism has no other author. */
    public Map<String, Object> readableContent() {
        StringBuilder full = new StringBuilder(text == null ? "" : text);
        for (String claim : claims) {
            full.append("\n\n").append(claim);
        }
        return Testimony.textContent(full.toString(), "the player");
    }

    public Snapshot snapshot() {
        return new Snapshot(text == null ? "" : text, claims);
    }

    public static Board restore(Snapshot snapshot) {
        Board board = new Board();
        board.text = snapshot.text().isEmpty() ? null : snapshot.text();
        board.claims.addAll(snapshot.claims());
        return board;
    }
}
