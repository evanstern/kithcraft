package dev.kithcraft.mod.board;

import dev.kithcraft.mod.percept.PerceptEmitter;
import dev.kithcraft.mod.percept.Provenance;
import dev.kithcraft.mod.percept.Testimony;
import org.junit.jupiter.api.Test;

import java.util.Map;
import java.util.Set;

import static org.junit.jupiter.api.Assertions.assertEquals;

/**
 * T002 (card AC #3): the board's read percept crosses using ONLY V2's existing §4.7 machinery
 * — no new percept field, no new content key, no new target shape. Proven structurally by
 * composing the exact content/provenance {@link BoardVisit}'s own read cycle fires and
 * asserting its shape is identical to {@link Testimony}'s pre-existing {@code text}
 * content/provenance — the same reuse {@code GraveBoardEntryTest} already proved for V5's
 * grave posting. (Phase 1 does not touch the act surface, so there is no new target shape to
 * check here — the claim intent's target is Phase 2's concern.)
 */
class BoardReadProtocolTest {

    @Test
    void readableContentHasExactlyTheExistingTextContentKeys() {
        Board board = new Board();
        board.post("build a wall");
        Map<String, Object> content = board.readableContent();
        assertEquals(Set.of("text", "attributed_to"), content.keySet(),
            "card AC #3: no new content key — the board's read content is exactly §4.7's shape");
    }

    @Test
    void composedEnvelopeIsAnOrdinaryTextReadPercept() {
        Board board = new Board();
        board.post("build a wall");
        Provenance.Source boardSource = Provenance.Source.artifact("th-board", "the job board");
        Provenance provenance = Testimony.textProvenance(boardSource, null, 100L);

        PerceptEmitter emitter = new PerceptEmitter(m -> { }, "s-board-test");
        Map<String, Object> envelope = emitter.compose(
            "b-tam", "text", PerceptEmitter.Urgency.NOTABLE, provenance, null, board.readableContent(), 100L);

        @SuppressWarnings("unchecked")
        Map<String, Object> payload = (Map<String, Object>) envelope.get("payload");
        assertEquals("text", payload.get("percept_type"));
        @SuppressWarnings("unchecked")
        Map<String, Object> prov = (Map<String, Object>) payload.get("provenance");
        assertEquals("read", prov.get("origin"));
        assertEquals(Set.of("percept_id", "percept_type", "urgency", "provenance", "place", "content"),
            payload.keySet(), "card AC #3: the envelope carries no field beyond §4.1's existing shape");
    }
}
