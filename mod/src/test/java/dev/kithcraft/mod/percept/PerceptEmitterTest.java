package dev.kithcraft.mod.percept;

import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T003: envelope shape, per-body monotonic seq assigned at enqueue, and §4.11's shedding
 * contract (urgent never shed; a shed background percept still leaves an honest seq gap). */
class PerceptEmitterTest {

    private static Map<String, Object> content() {
        return Sightings.sightingContent(
            Thing.of(null, "k:sleeping-place", List.of("shelter"), "a bed"), "near", null);
    }

    @Test
    void envelopeCarriesTheRequiredEnvelopeFields() throws Exception {
        List<Object> sent = new ArrayList<>();
        PerceptEmitter emitter = new PerceptEmitter(sent::add, "s-7f3a");
        Provenance prov = Provenance.direct(Provenance.Origin.SAW, 100L, 100L);
        emitter.emit("b-eda", "sighting", PerceptEmitter.Urgency.BACKGROUND, prov,
            new Place("pl-1", "the well"), content(), 918233L);

        assertEquals(1, sent.size());
        @SuppressWarnings("unchecked")
        Map<String, Object> envelope = (Map<String, Object>) sent.get(0);
        assertEquals("0.1", envelope.get("protocol"));
        assertEquals("percept", envelope.get("message"));
        assertEquals("s-7f3a", envelope.get("session"));
        assertEquals(1L, envelope.get("seq"));
        assertEquals("b-eda", envelope.get("body"));
        assertEquals(918233L, envelope.get("world_time"));
        assertTrue(envelope.get("payload") instanceof Map);
    }

    @Test
    void seqIsMonotonicPerBodyIndependently() throws Exception {
        List<Object> sent = new ArrayList<>();
        PerceptEmitter emitter = new PerceptEmitter(sent::add, "s-1");
        Provenance prov = Provenance.direct(Provenance.Origin.SAW, 1L, 1L);
        emitter.emit("b-a", "sighting", PerceptEmitter.Urgency.BACKGROUND, prov, null, content(), 1L);
        emitter.emit("b-a", "sighting", PerceptEmitter.Urgency.BACKGROUND, prov, null, content(), 2L);
        emitter.emit("b-b", "sighting", PerceptEmitter.Urgency.BACKGROUND, prov, null, content(), 1L);

        assertEquals(1L, envelopeAt(sent, 0).get("seq"));
        assertEquals(2L, envelopeAt(sent, 1).get("seq"));
        assertEquals(1L, envelopeAt(sent, 2).get("seq"), "body b starts its own counter at 1");
    }

    @Test
    void shedBackgroundPerceptLeavesAnHonestSeqGap() throws Exception {
        List<Object> sent = new ArrayList<>();
        PerceptEmitter emitter = new PerceptEmitter(sent::add, "s-1");
        Provenance prov = Provenance.direct(Provenance.Origin.SAW, 1L, 1L);

        Map<String, Object> shed = emitter.emitOrShed(
            "b-a", "sighting", PerceptEmitter.Urgency.BACKGROUND, prov, null, content(), 1L, true);
        emitter.emit("b-a", "sighting", PerceptEmitter.Urgency.BACKGROUND, prov, null, content(), 2L);

        assertNull(shed, "shed percepts are not sent");
        assertEquals(1, sent.size());
        // seq 1 was consumed by the shed percept (seq-at-enqueue); the delivered one is 2 —
        // the gap IS the shedding signal (§4.11).
        assertEquals(2L, envelopeAt(sent, 0).get("seq"));
    }

    @Test
    void urgentPerceptsMustNotBeShed() {
        PerceptEmitter emitter = new PerceptEmitter(m -> {}, "s-1");
        Provenance prov = Provenance.direct(Provenance.Origin.FELT, 1L, 1L);
        assertThrows(IllegalArgumentException.class, () -> emitter.emitOrShed(
            "b-a", "self_state", PerceptEmitter.Urgency.URGENT, prov, null,
            SelfState.content(SelfState.COLD, SelfState.Level.CRITICAL, SelfState.Trend.WORSENING), 1L, true));
    }

    @SuppressWarnings("unchecked")
    private static Map<String, Object> envelopeAt(List<Object> sent, int i) {
        return (Map<String, Object>) sent.get(i);
    }
}
