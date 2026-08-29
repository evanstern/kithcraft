package dev.kithcraft.mod.board;

import org.junit.jupiter.api.Test;

import java.util.Map;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T001/T003: free text is a valid posting with no form or syntax (card ACs #1, #7); claims
 * appended to the readable content (card AC #4's read half). T004/T006: {@link
 * Board#tryClaim} itself (the engine-side claim gate {@link BoardClaims} calls) and {@link
 * Board#postNotice} (V5's tend-grave seam closure). */
class BoardTest {

    @Test
    void freeTextIsAValidPostingNoFormOrSyntax() {
        Board board = new Board();
        board.post("build a wall. maybe stone? whatever's around");
        Map<String, Object> content = board.readableContent();
        assertEquals("build a wall. maybe stone? whatever's around", content.get("text"));
        assertEquals("the player", content.get("attributed_to"));
    }

    @Test
    void emptyBoardStillReadsCleanly() {
        Board board = new Board();
        assertEquals("", board.readableContent().get("text"));
    }

    @Test
    void claimsAreAppendedToTheReadableContent() {
        Board board = new Board();
        board.post("dig a well");
        board.recordClaim("Claimed by Tam.");
        String text = (String) board.readableContent().get("text");
        assertTrue(text.contains("dig a well"));
        assertTrue(text.contains("Claimed by Tam."));
    }

    @Test
    void snapshotRoundTripsTextAndClaims() {
        Board board = new Board();
        board.post("dig a well");
        board.recordClaim("Claimed by Tam.");
        Board restored = Board.restore(board.snapshot());
        assertEquals(board.readableContent(), restored.readableContent());
    }

    @Test
    void freshBoardSnapshotRestoresWithNoPosting() {
        Board restored = Board.restore(new Board().snapshot());
        assertEquals("", restored.readableContent().get("text"));
    }

    @Test
    void tryClaimFirstAcceptedClaimWins() {
        Board board = new Board();
        board.post("dig a well");
        assertTrue(board.tryClaim("b-tam", "Tam"));
        assertEquals(Optional.of("b-tam"), board.claimedBy());
        assertTrue(((String) board.readableContent().get("text")).contains("Tam has claimed this."));
    }

    @Test
    void tryClaimSameClaimantAgainIsIdempotent() {
        Board board = new Board();
        assertTrue(board.tryClaim("b-tam", "Tam"));
        assertTrue(board.tryClaim("b-tam", "Tam"), "the same claimant claiming again does not lose");
        assertEquals(1, board.claims().size(), "no duplicate claim line for the same claimant");
    }

    @Test
    void tryClaimADifferentClaimantAfterAlreadyClaimedLoses() {
        Board board = new Board();
        assertTrue(board.tryClaim("b-tam", "Tam"));
        assertFalse(board.tryClaim("b-mira", "Mira"), "first accepted claim wins (spec Edge Cases)");
        assertEquals(Optional.of("b-tam"), board.claimedBy());
        assertFalse(((String) board.readableContent().get("text")).contains("Mira has claimed this."));
    }

    @Test
    void postNoticeRidesTheSameReadableContentAsAClaim() {
        Board board = new Board();
        board.post("dig a well");
        board.postNotice("Tend Tam's grave.");
        String text = (String) board.readableContent().get("text");
        assertTrue(text.contains("dig a well"));
        assertTrue(text.contains("Tend Tam's grave."), "T006: a grave posting rides this board for real");
    }

    @Test
    void snapshotRoundTripsClaimState() {
        Board board = new Board();
        board.tryClaim("b-tam", "Tam");
        Board restored = Board.restore(board.snapshot());
        assertEquals(Optional.of("b-tam"), restored.claimedBy());
        assertFalse(restored.tryClaim("b-mira", "Mira"), "claim state survives a restart");
    }
}
