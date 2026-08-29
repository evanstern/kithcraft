package dev.kithcraft.mod.board;

import dev.kithcraft.mod.act.ClaimRegistry;
import dev.kithcraft.mod.act.IntentHandler;
import dev.kithcraft.mod.act.TargetResolution;
import dev.kithcraft.mod.act.Verbs;
import dev.kithcraft.mod.percept.PerceptEmitter;
import dev.kithcraft.mod.percept.Place;
import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * T004/T005 (card AC #4, #9): a claim intent driven through the REAL engine-side pipeline —
 * {@link IntentHandler} → {@link BoardClaims} → {@link Board#tryClaim} — and card AC #4's
 * "appears in board content" checked at the end, not asserted in isolation.
 *
 * <p><b>T005, recorded honestly:</b> the intent payload below is an M5-shaped SCRIPTED
 * intent, not a generated cross-language fixture. mind/deliberate has no fixture-export path
 * (its tests script raw JSON strings inline, e.g. board_test.go's {@code
 * TestLoop_E3ClaimIntent_CarriesAuthoredReason}); building one was judged out of this
 * phase's scope. Instead this payload is asserted, field-for-field, against the two real Go
 * shapes that produce it: {@code mind/llm/structured.go}'s {@code Intent} struct ({@code
 * verb}, {@code target}, {@code reason}, {@code supersedes} — exactly what {@code
 * llm.ParseIntent} decodes) composed through {@code mind/seam/intents.go}'s {@code
 * Pending.Compose}, which adds {@code intent_id} and {@code not_after: nil}. The verb, target
 * shape, and reason text below are copied verbatim from board_test.go's own scripted E3
 * output so a future reader can diff this against the real Go test directly.
 */
class BoardClaimTest {
    // Copied verbatim from mind/deliberate/board_test.go's
    // TestLoop_E3ClaimIntent_CarriesAuthoredReason — the real M5 loop's own scripted output.
    private static final String M5_REASON =
        "Building shelters is exactly my trade, and that north wall has bothered me all week.";

    /** Exactly seam.Pending.Compose's payload shape (intents.go): {@code intent_id, verb,
     * target, reason, supersedes, not_after} — nothing else. */
    private static Map<String, Object> m5ShapedClaimIntent(String intentId, String thingToken, String reason) {
        Map<String, Object> payload = new java.util.LinkedHashMap<>();
        payload.put("intent_id", intentId);
        payload.put("verb", "claim");
        payload.put("target", Map.of("type", "thing", "thing_id", thingToken));
        payload.put("reason", reason);
        payload.put("supersedes", null);
        payload.put("not_after", null);
        return payload;
    }

    private static TargetResolution.Ground groundKnowing(String... issuedThingTokens) {
        return new TargetResolution.Ground() {
            @Override public boolean isIssued(String token) {
                return List.of(issuedThingTokens).contains(token);
            }
            @Override public Optional<Place> placeFor(String token) { return Optional.empty(); }
        };
    }

    @SuppressWarnings("unchecked")
    private static Map<String, Object> contentOf(Object envelope) {
        Map<String, Object> payload = (Map<String, Object>) ((Map<String, Object>) envelope).get("payload");
        return (Map<String, Object>) payload.get("content");
    }

    @Test
    void m5ShapedClaimIntentRegistersAndAppearsInBoardContent() throws Exception {
        Board board = new Board();
        board.post("build a wall by the north edge");
        BoardClaims claims = new BoardClaims(board, "th-board", token -> Optional.of("Tam"));

        List<Object> sent = new ArrayList<>();
        IntentHandler handler = new IntentHandler(new PerceptEmitter(sent::add, "s-1"),
            groundKnowing("th-board"), noWalking(), claims, "s-1", "b-tam");

        Map<String, Object> ack = handler.handle(m5ShapedClaimIntent("i-1", "th-board", M5_REASON), 100L);
        assertTrue((Boolean) ((Map<String, Object>) ack.get("payload")).get("accepted"));

        Map<String, Object> content = contentOf(sent.get(0));
        assertEquals("completed", content.get("outcome"));
        assertEquals(M5_REASON, content.get("reason"), "the mind's own authored reason, echoed unaltered (§5.2)");

        // Card AC #4's read half, closed for real: the claim is now visible on the board.
        String boardText = (String) board.readableContent().get("text");
        assertTrue(boardText.contains("Tam has claimed this."),
            "card AC #4: a registered claim appears in the board's readable content");
    }

    @Test
    void secondClaimantLosesAndTheLoserSeesItOnTheirNextRead() throws Exception {
        Board board = new Board();
        board.post("dig a well");
        BoardClaims claims = new BoardClaims(board, "th-board",
            token -> Optional.of("b-tam".equals(token) ? "Tam" : "Mira"));

        List<Object> sentTam = new ArrayList<>();
        IntentHandler tam = new IntentHandler(new PerceptEmitter(sentTam::add, "s-1"),
            groundKnowing("th-board"), noWalking(), claims, "s-1", "b-tam");
        List<Object> sentMira = new ArrayList<>();
        IntentHandler mira = new IntentHandler(new PerceptEmitter(sentMira::add, "s-1"),
            groundKnowing("th-board"), noWalking(), claims, "s-1", "b-mira");

        tam.handle(m5ShapedClaimIntent("i-1", "th-board", "I've done this before"), 100L);
        mira.handle(m5ShapedClaimIntent("i-2", "th-board", "I was closer"), 101L);

        assertEquals("completed", contentOf(sentTam.get(0)).get("outcome"));
        Map<String, Object> miraResult = contentOf(sentMira.get(0));
        assertEquals("failed", miraResult.get("outcome"));
        assertEquals("blocked", miraResult.get("reason_code"));

        // Spec Edge Cases: "the loser's next read shows the claim" — the board's content
        // proves it, not just the act_result.
        String boardText = (String) board.readableContent().get("text");
        assertTrue(boardText.contains("Tam has claimed this."));
        assertFalse(boardText.contains("Mira has claimed this."));
    }

    private static Verbs.Actuator noWalking() {
        return new Verbs.Actuator() {
            @Override public Verbs.WalkStatus stepWalk(String body, Place place) {
                throw new AssertionError("claim never walks");
            }
            @Override public void speak(String body, Map<String, Object> target, String text) { }
            @Override public void attend(String body, Place place) { }
        };
    }
}
