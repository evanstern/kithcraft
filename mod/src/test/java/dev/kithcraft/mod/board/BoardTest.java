package dev.kithcraft.mod.board;

import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T001/T003: free text is a valid posting with no form or syntax (card ACs #1, #7); claims
 * appended to the readable content (card AC #4's read half). */
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
}
