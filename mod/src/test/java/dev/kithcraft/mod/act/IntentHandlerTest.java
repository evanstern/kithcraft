package dev.kithcraft.mod.act;

import dev.kithcraft.mod.percept.PerceptEmitter;
import dev.kithcraft.mod.percept.Place;
import org.junit.jupiter.api.Test;

import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T007 (card AC #6): intent decode/ack/act_result state machine, and §5.7 cancel. */
class IntentHandlerTest {

    private static TargetResolution.Ground stubGround(Map<String, Place> issued) {
        return new TargetResolution.Ground() {
            @Override public boolean isIssued(String token) {
                return issued.containsKey(token);
            }
            @Override public Optional<Place> placeFor(String token) {
                return Optional.ofNullable(issued.get(token));
            }
        };
    }

    /** An actuator that arrives on the first {@code stepWalk} call — enough to exercise
     * speak/attend/wait/immediate-go_to paths without a multi-tick loop. */
    private static final class ImmediateActuator implements Verbs.Actuator {
        final List<Object[]> speaks = new ArrayList<>();
        final List<Place> attends = new ArrayList<>();
        @Override public Verbs.WalkStatus stepWalk(String body, Place place) { return Verbs.WalkStatus.ARRIVED; }
        @Override public void speak(String body, Map<String, Object> target, String text) { speaks.add(new Object[]{body, target, text}); }
        @Override public void attend(String body, Place place) { attends.add(place); }
    }

    /** An actuator that returns a scripted sequence of {@link Verbs.WalkStatus} values, one
     * per {@code stepWalk} call — drives the multi-tick cases (AC #8). */
    private static final class ScriptedActuator implements Verbs.Actuator {
        private final Deque<Verbs.WalkStatus> script;
        ScriptedActuator(Verbs.WalkStatus... steps) { this.script = new ArrayDeque<>(List.of(steps)); }
        @Override public Verbs.WalkStatus stepWalk(String body, Place place) { return script.poll(); }
        @Override public void speak(String body, Map<String, Object> target, String text) {}
        @Override public void attend(String body, Place place) {}
    }

    @SuppressWarnings("unchecked")
    private static Map<String, Object> payloadOf(Object envelope) {
        return (Map<String, Object>) ((Map<String, Object>) envelope).get("payload");
    }

    @SuppressWarnings("unchecked")
    private static List<Map<String, Object>> actResultsOf(List<Object> sent) {
        List<Map<String, Object>> results = new ArrayList<>();
        for (Object o : sent) {
            Map<String, Object> env = (Map<String, Object>) o;
            if ("percept".equals(env.get("message"))
                    && "act_result".equals(((Map<String, Object>) env.get("payload")).get("percept_type"))) {
                results.add(env);
            }
        }
        return results;
    }

    @Test
    void unknownVerbIsRefusedAndProducesNoActResult() throws Exception {
        List<Object> sent = new ArrayList<>();
        IntentHandler handler = new IntentHandler(new PerceptEmitter(sent::add, "s-1"),
            stubGround(Map.of()), new ImmediateActuator(), "s-1", "b-eda");

        Map<String, Object> ack = handler.handle(
            Map.of("intent_id", "i-1", "verb", "flurbosplat"), 100L);

        assertFalse((Boolean) payloadOf(ack).get("accepted"));
        assertEquals("unknown_verb", payloadOf(ack).get("reason_code"));
        assertTrue(sent.isEmpty(), "a refused intent owes no act_result");
    }

    @Test
    void unknownTargetIsRefusedForAnUnissuedToken() throws Exception {
        List<Object> sent = new ArrayList<>();
        IntentHandler handler = new IntentHandler(new PerceptEmitter(sent::add, "s-1"),
            stubGround(Map.of()), new ImmediateActuator(), "s-1", "b-eda");

        Map<String, Object> ack = handler.handle(
            Map.of("intent_id", "i-1", "verb", "go_to", "target", Map.of("type", "body", "body", "b-ghost")),
            100L);

        assertFalse((Boolean) payloadOf(ack).get("accepted"));
        assertEquals("unknown_target", payloadOf(ack).get("reason_code"));
        assertTrue(sent.isEmpty());
    }

    @Test
    void everyCoreVerbAckedAcceptedYieldsExactlyOneActResult() throws Exception {
        for (Map<String, Object> intent : List.of(
                Map.of("intent_id", "i-go", "verb", "go_to", "target", Map.of("type", "place", "place", "pl-1")),
                Map.of("intent_id", "i-sp", "verb", "speak", "target", Map.of("type", "none")),
                Map.of("intent_id", "i-at", "verb", "attend", "target", Map.of("type", "none")),
                Map.of("intent_id", "i-wa", "verb", "wait", "target", Map.of("type", "none")))) {
            List<Object> sent = new ArrayList<>();
            IntentHandler handler = new IntentHandler(new PerceptEmitter(sent::add, "s-1"),
                stubGround(Map.of("pl-1", new Place("pl-1", "the well"))), new ImmediateActuator(), "s-1", "b-eda");

            Map<String, Object> ack = handler.handle(intent, 100L);

            assertTrue((Boolean) payloadOf(ack).get("accepted"), intent.get("verb") + " should be accepted");
            List<Map<String, Object>> results = actResultsOf(sent);
            assertEquals(1, results.size(), intent.get("verb") + " must yield exactly one act_result (card AC #6)");
            Map<String, Object> content = (Map<String, Object>) payloadOf(results.get(0)).get("content");
            assertEquals(intent.get("intent_id"), content.get("intent_id"));
            assertEquals("completed", content.get("outcome"));
        }
    }

    /** Card AC #7 at the handler level: go_to targeting a body resolves and walks to that
     * body's last-seen place — the actuator only ever sees the place {@link
     * TargetResolution.Ground} supplied, never anything else. */
    @Test
    void goToABodyWalksToItsLastSeenPlace() throws Exception {
        List<Object> sent = new ArrayList<>();
        Place lastSeen = new Place("pl-51c", "the old orchard");
        List<Place> walkedTo = new ArrayList<>();
        Verbs.Actuator actuator = new Verbs.Actuator() {
            @Override public Verbs.WalkStatus stepWalk(String body, Place place) { walkedTo.add(place); return Verbs.WalkStatus.ARRIVED; }
            @Override public void speak(String body, Map<String, Object> target, String text) {}
            @Override public void attend(String body, Place place) {}
        };
        IntentHandler handler = new IntentHandler(new PerceptEmitter(sent::add, "s-1"),
            stubGround(Map.of("b-tam", lastSeen)), actuator, "s-1", "b-eda");

        handler.handle(Map.of("intent_id", "i-1", "verb", "go_to", "target",
            Map.of("type", "body", "body", "b-tam")), 100L);

        assertEquals(List.of(lastSeen), walkedTo);
    }

    /** Card AC #8: a known-but-gone body target is ACCEPTED, the walk proceeds, and only
     * after it fails does exactly one act_result arrive with target_gone — never at ack time. */
    @Test
    void knownButGoneReferentAcceptsThenFailsTargetGoneAfterTheWalk() throws Exception {
        List<Object> sent = new ArrayList<>();
        Place lastSeen = new Place("pl-51c", "the old orchard");
        ScriptedActuator actuator = new ScriptedActuator(Verbs.WalkStatus.IN_PROGRESS, Verbs.WalkStatus.TARGET_GONE);
        IntentHandler handler = new IntentHandler(new PerceptEmitter(sent::add, "s-1"),
            stubGround(Map.of("b-tam", lastSeen)), actuator, "s-1", "b-eda");

        Map<String, Object> ack = handler.handle(Map.of("intent_id", "i-1", "verb", "go_to", "target",
            Map.of("type", "body", "body", "b-tam")), 100L);

        assertTrue((Boolean) payloadOf(ack).get("accepted"), "known-but-gone is accepted, not refused (§5.6)");
        assertTrue(actResultsOf(sent).isEmpty(), "no act_result yet — the walk is still in progress");

        handler.tick(101L);

        List<Map<String, Object>> results = actResultsOf(sent);
        assertEquals(1, results.size(), "exactly one act_result, after the walk (card AC #6/#8)");
        Map<String, Object> content = (Map<String, Object>) payloadOf(results.get(0)).get("content");
        assertEquals("failed", content.get("outcome"));
        assertEquals("target_gone", content.get("reason_code"));
    }

    @Test
    void cancelOfAPendingWalkProducesASupersededActResult() throws Exception {
        List<Object> sent = new ArrayList<>();
        Place place = new Place("pl-1", "the well");
        ScriptedActuator actuator = new ScriptedActuator(Verbs.WalkStatus.IN_PROGRESS);
        IntentHandler handler = new IntentHandler(new PerceptEmitter(sent::add, "s-1"),
            stubGround(Map.of("pl-1", place)), actuator, "s-1", "b-eda");
        handler.handle(Map.of("intent_id", "i-1", "verb", "go_to", "target", Map.of("type", "place", "place", "pl-1")), 100L);
        assertTrue(actResultsOf(sent).isEmpty());

        handler.cancel("i-1", 105L);

        List<Map<String, Object>> results = actResultsOf(sent);
        assertEquals(1, results.size());
        Map<String, Object> content = (Map<String, Object>) payloadOf(results.get(0)).get("content");
        assertEquals("superseded", content.get("outcome"));
    }

    @Test
    void supersedesCancelsThePriorPendingIntentBeforeAcceptingTheNewOne() throws Exception {
        List<Object> sent = new ArrayList<>();
        Place place = new Place("pl-1", "the well");
        ScriptedActuator actuator = new ScriptedActuator(Verbs.WalkStatus.IN_PROGRESS, Verbs.WalkStatus.IN_PROGRESS);
        IntentHandler handler = new IntentHandler(new PerceptEmitter(sent::add, "s-1"),
            stubGround(Map.of("pl-1", place)), actuator, "s-1", "b-eda");
        handler.handle(Map.of("intent_id", "i-1", "verb", "go_to", "target", Map.of("type", "place", "place", "pl-1")), 100L);

        handler.handle(Map.of("intent_id", "i-2", "verb", "go_to", "target", Map.of("type", "place", "place", "pl-1"),
            "supersedes", "i-1"), 101L);

        List<Map<String, Object>> results = actResultsOf(sent);
        assertEquals(1, results.size(), "i-1 gets its superseded result immediately; i-2 is still walking");
        assertEquals("i-1", ((Map<String, Object>) payloadOf(results.get(0)).get("content")).get("intent_id"));
        assertEquals("superseded", ((Map<String, Object>) payloadOf(results.get(0)).get("content")).get("outcome"));
    }

    @Test
    void notAfterExpiryAbandonsAPendingWalkAsExpired() throws Exception {
        List<Object> sent = new ArrayList<>();
        Place place = new Place("pl-1", "the well");
        ScriptedActuator actuator = new ScriptedActuator(Verbs.WalkStatus.IN_PROGRESS);
        IntentHandler handler = new IntentHandler(new PerceptEmitter(sent::add, "s-1"),
            stubGround(Map.of("pl-1", place)), actuator, "s-1", "b-eda");
        handler.handle(Map.of("intent_id", "i-1", "verb", "go_to", "target", Map.of("type", "place", "place", "pl-1"),
            "not_after", 110L), 100L);

        handler.tick(120L);

        List<Map<String, Object>> results = actResultsOf(sent);
        assertEquals(1, results.size());
        Map<String, Object> content = (Map<String, Object>) payloadOf(results.get(0)).get("content");
        assertEquals("expired", content.get("outcome"));
        assertNull(content.get("reason_code"));
    }

    @Test
    void issuedTokenWithNoRecordedPlaceFailsUnreachableNotTargetGone() throws Exception {
        List<Object> sent = new ArrayList<>();
        Map<String, Place> issuedNoPlace = new java.util.HashMap<>();
        issuedNoPlace.put("b-tam", null);
        IntentHandler handler = new IntentHandler(new PerceptEmitter(sent::add, "s-1"),
            stubGround(issuedNoPlace), new ImmediateActuator(), "s-1", "b-eda");

        Map<String, Object> ack = handler.handle(Map.of("intent_id", "i-1", "verb", "go_to", "target",
            Map.of("type", "body", "body", "b-tam")), 100L);

        assertTrue((Boolean) payloadOf(ack).get("accepted"));
        Map<String, Object> content = (Map<String, Object>) payloadOf(actResultsOf(sent).get(0)).get("content");
        assertEquals("failed", content.get("outcome"));
        assertEquals("unreachable", content.get("reason_code"));
    }
}
