package dev.kithcraft.mod.percept;

import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T002/T003: sighting vs. observation shape, card AC #2's absence-claim scope, first-
 * sighting tracking, dedup, and the free nearest_hostile danger signal. */
class SightingsTest {

    @Test
    void sightingIsABoundedClaimAboutOneThing() {
        Thing tam = new Thing("th-401", "k:person", List.of(), "Tam", "b-tam", 1);
        Map<String, Object> content = Sightings.sightingContent(tam, "near", "drawing water");
        assertEquals("near", content.get("distance"));
        assertEquals("drawing water", content.get("doing"));
        assertEquals("k:person", ((Map<?, ?>) content.get("thing")).get("kind"));
    }

    @Test
    void observationVocabularyMustBeDeclaredInTheManifest() {
        assertThrows(IllegalArgumentException.class,
            () -> Sightings.observationContent("near", List.of("k:not-a-declared-kind"), List.of()));
    }

    @Test
    void observationRejectsAPresentThingOutsideItsOwnVocabulary() {
        Thing hostile = Thing.of(null, "k:hostile", List.of("danger"), "a zombie");
        assertThrows(IllegalArgumentException.class,
            () -> Sightings.observationContent("near", List.of("k:person"), List.of(hostile)));
    }

    /** Card AC #2: an observation yields a falsifiable absence claim scoped exactly to
     * {@code vocabulary} — nothing outside the scanned vocabulary can be asserted absent,
     * and everything in the vocabulary not present IS the absence claim. */
    @Test
    void absenceClaimIsScopedExactlyToTheScannedVocabulary() {
        List<String> vocabulary = List.of("k:sleeping-place", "k:person", "k:hostile", "k:door");
        Thing person = Thing.of(null, "k:person", List.of(), "someone");
        List<Thing> present = List.of(person);

        Set<String> absent = Sightings.absentKinds(vocabulary, present);

        assertEquals(Set.of("k:sleeping-place", "k:hostile", "k:door"), absent,
            "absence is vocabulary minus present, and nothing else (§4.3)");
        assertFalse(absent.contains("k:person"), "a present kind is never claimed absent");
        // A kind never in the vocabulary at all is out of scope for the absence claim
        // entirely — it is neither present nor absent, it was simply never scanned.
        assertFalse(vocabulary.contains("k:workbench"));
    }

    @Test
    void firstSightingIsTrueOnceThenFalse() {
        Sightings s = new Sightings();
        assertTrue(s.isFirstSighting("b-eda", "th-401"));
        assertFalse(s.isFirstSighting("b-eda", "th-401"));
        assertTrue(s.isFirstSighting("b-tam", "th-401"), "tracked per-body, not globally");
    }

    @Test
    void observationChangedDetectsAnUnchangedPlaceForDedup() {
        Sightings s = new Sightings();
        Thing bed = Thing.of(null, "k:sleeping-place", List.of("shelter"), "a bed");
        assertTrue(s.observationChanged("b-eda", "pl-1", List.of(bed)), "first observation always counts as changed");
        assertFalse(s.observationChanged("b-eda", "pl-1", List.of(bed)), "unchanged present-set is suppressed (§4.3)");
        Thing door = Thing.of(null, "k:door", List.of("boundary"), "a door");
        assertTrue(s.observationChanged("b-eda", "pl-1", List.of(bed, door)), "a real change is never suppressed");
    }

    @Test
    void nearestHostileRidesAnOrdinarySightingOfTheDeclaredHostileKind() {
        Sightings.HostileSource source = body -> Optional.of(new Sightings.HostileSource.Nearby("th-99", "near"));
        Map<String, Object> content = Sightings.nearestHostileSighting(source, "b-eda").orElseThrow();
        Map<?, ?> thing = (Map<?, ?>) content.get("thing");
        assertEquals("k:hostile", thing.get("kind"));
        assertEquals(List.of("danger"), thing.get("roles"));
        assertEquals("near", content.get("distance"));
    }

    @Test
    void noHostileYieldsNoSighting() {
        Sightings.HostileSource none = body -> Optional.empty();
        assertTrue(Sightings.nearestHostileSighting(none, "b-eda").isEmpty());
    }
}
