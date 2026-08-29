package dev.kithcraft.mod.board;

import dev.kithcraft.mod.percept.Testimony;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;

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
    private String claimedBy;

    /** Snapshot for {@link BoardData}'s persistence — a plain value, no Minecraft types,
     * mirroring {@code TokenRegistry.Snapshot}'s own shape. {@code claimedBy} is {@code ""}
     * for unclaimed, the same empty-string-means-absent convention {@code text} already uses. */
    public record Snapshot(String text, List<String> claims, String claimedBy) {
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

    /** T004 (card AC #4): appends one claim line to the board's visible content. This class
     * only carries lines — it does not dedupe or validate them; {@link #tryClaim} is the
     * real claim-identity gate that calls this. */
    public void recordClaim(String claimLine) {
        claims.add(Objects.requireNonNull(claimLine, "claimLine"));
    }

    /** T006 (V5 seam closure): appends an arbitrary board notice — {@code
     * dev.kithcraft.mod.death.GraveBoardEntry}'s posting riding this board for real, closing
     * the decoupling that class's own doc flagged ("when V4 merges its board book, the same
     * content rides that channel with no rework"). Same list as {@link #recordClaim}; a
     * distinct method name only so a non-claim call site (death handling) reads honestly. */
    public void postNotice(String text) {
        recordClaim(text);
    }

    /** T004 (card AC #4, AR-4): registers the FIRST accepted claim on the current posting —
     * AR-4's "the mind names a token, the engine resolves it" lands here. Idempotent for the
     * SAME claimant; a DIFFERENT claimant while already held loses and this returns {@code
     * false} (spec Edge Cases: "first accepted claim wins engine-side; the loser's next read
     * shows the claim"). {@code claimantDescriptor} composes the board-visible line — never
     * the raw body token — so a claim reads like a person claimed it. Package-private: the
     * ONLY caller is {@link BoardClaims}, resolving a claim intent (card AC #9 — no other
     * path, including any player-facing one, may call this; {@code StructuralAbsenceTest}
     * checks the call-site count structurally). */
    synchronized boolean tryClaim(String claimantBody, String claimantDescriptor) {
        Objects.requireNonNull(claimantBody, "claimantBody");
        if (claimedBy != null) {
            return claimedBy.equals(claimantBody);
        }
        claimedBy = claimantBody;
        recordClaim(claimantDescriptor + " has claimed this.");
        return true;
    }

    /** The current claimant's body token, or empty if unclaimed. */
    public synchronized Optional<String> claimedBy() {
        return Optional.ofNullable(claimedBy);
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
        return new Snapshot(text == null ? "" : text, claims, claimedBy == null ? "" : claimedBy);
    }

    public static Board restore(Snapshot snapshot) {
        Board board = new Board();
        board.text = snapshot.text().isEmpty() ? null : snapshot.text();
        board.claims.addAll(snapshot.claims());
        board.claimedBy = snapshot.claimedBy().isEmpty() ? null : snapshot.claimedBy();
        return board;
    }
}
