package dev.kithcraft.mod.percept;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestFactory;
import org.junit.jupiter.api.DynamicTest;

import java.util.ArrayDeque;
import java.util.Deque;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * T003: every emitted shape from this phase (sighting, observation, self_state) passes
 * AR-1..AR-6, and card AC #4's no-salience-field proof holds over each — a reflective
 * recursive scan, not a spot check, so a future field addition cannot slip past it silently.
 */
class AbstractionConformanceTest {

    /** Card AC #4: no salience/importance/weight/memorability field may exist anywhere
     * (§2.8). */
    private static final Set<String> FORBIDDEN_KEYS =
        Set.of("salience", "importance", "weight", "memorability");

    /** AR-4: no coordinates, ever. */
    private static final Set<String> COORDINATE_KEYS = Set.of("x", "y", "z", "pos", "position");

    private static final Set<String> DISTANCE_BANDS = Set.of("here", "near", "middling", "far");
    private static final Set<String> CONDITION_LEVELS = Set.of("none", "mild", "severe", "critical");

    /** Recursively asserts AR-1/AR-2/AR-4/AR-6 and the no-salience-field proof over an
     * arbitrary wire-shaped value (Map/List/String/Long/Boolean/null — CanonicalJson's own
     * value set). */
    static void assertConformant(Object value) {
        Deque<Object> stack = new ArrayDeque<>();
        stack.push(value);
        while (!stack.isEmpty()) {
            Object v = stack.pop();
            if (v instanceof Map<?, ?> m) {
                for (Map.Entry<?, ?> e : m.entrySet()) {
                    String key = (String) e.getKey();
                    assertTrue(!FORBIDDEN_KEYS.contains(key),
                        "forbidden salience-shaped field \"" + key + "\" (§2.8, card AC #4)");
                    assertTrue(!COORDINATE_KEYS.contains(key),
                        "forbidden coordinate field \"" + key + "\" (AR-4)");
                    if ("kind".equals(key) && e.getValue() instanceof String kind) {
                        assertTrue(!kind.contains(":") || kind.startsWith("k:"),
                            "kind token \"" + kind + "\" is not an opaque vendor token (AR-1/AR-2)");
                        assertTrue(!kind.toLowerCase(java.util.Locale.ROOT).startsWith("minecraft"),
                            "kind token \"" + kind + "\" leaks a native namespace (AR-1)");
                    }
                    if (("distance".equals(key) || "extent".equals(key) || "bearing".equals(key))
                            && e.getValue() != null) {
                        assertTrue(DISTANCE_BANDS.contains(e.getValue())
                                || "ahead".equals(e.getValue()) || "behind".equals(e.getValue())
                                || "left".equals(e.getValue()) || "right".equals(e.getValue()),
                            key + " must be a coarse band, not a raw unit (AR-4): " + e.getValue());
                    }
                    if ("level".equals(key) && e.getValue() != null) {
                        assertTrue(CONDITION_LEVELS.contains(e.getValue()),
                            "level must be an AR-6 band, not a raw stat: " + e.getValue());
                    }
                    if (e.getValue() != null) {
                        stack.push(e.getValue());
                    }
                }
            } else if (v instanceof List<?> l) {
                l.forEach(item -> {
                    if (item != null) {
                        stack.push(item);
                    }
                });
            }
        }
    }

    @TestFactory
    Stream<DynamicTest> everyEmittedShapeIsConformant() {
        Thing bed = Thing.of(null, "k:sleeping-place", List.of("shelter", "sleeping_place"), "a bed");
        Thing tam = new Thing("th-401", "k:person", List.of(), "Tam", "b-tam", 1);

        Map<String, Object> sighting = Sightings.sightingContent(tam, "near", "drawing water");
        Map<String, Object> observation = Sightings.observationContent(
            "near", List.of("k:sleeping-place", "k:person", "k:hostile"), List.of(bed));
        Map<String, Object> selfState = SelfState.content(SelfState.COLD, SelfState.Level.SEVERE, SelfState.Trend.WORSENING);
        Map<String, Object> hostileSighting = Sightings.nearestHostileSighting(
                body -> java.util.Optional.of(new Sightings.HostileSource.Nearby("th-99", "near")), "b-eda")
            .orElseThrow();

        return Stream.of(
            DynamicTest.dynamicTest("sighting", () -> assertConformant(sighting)),
            DynamicTest.dynamicTest("observation", () -> assertConformant(observation)),
            DynamicTest.dynamicTest("self_state", () -> assertConformant(selfState)),
            DynamicTest.dynamicTest("nearest_hostile sighting", () -> assertConformant(hostileSighting)));
    }

    @Test
    void fullEnvelopeIsConformant() {
        PerceptEmitter.Sink recorder = m -> {};
        PerceptEmitter emitter = new PerceptEmitter(recorder, "s-1");
        Thing bed = Thing.of(null, "k:sleeping-place", List.of("shelter"), "a bed");
        Map<String, Object> content = Sightings.sightingContent(bed, "near", null);
        Provenance prov = Provenance.direct(Provenance.Origin.SAW, 100L, 100L);
        Map<String, Object> envelope = emitter.compose("b-eda", "sighting",
            PerceptEmitter.Urgency.BACKGROUND, prov, new Place("pl-1", "the well"), content, 100L);
        assertConformant(envelope);
        assertEquals("sighting", ((Map<?, ?>) envelope.get("payload")).get("percept_type"));
    }
}
