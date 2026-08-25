package dev.kithcraft.mod.percept;

import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T005 (§4.6-4.7): told_fact carries the teller's last-seen time, not the telling time
 * (EH-5); text is read-origin from an artifact and its attribution is not a told provenance. */
class TestimonyTest {

    @Test
    void toldFactObservedAtIsTellersLastSeenTimeNotReceivedAt() {
        Provenance.Source tam = Provenance.Source.body("b-tam", "Tam");
        long tellerLastSeenAt = 902_430L;
        long receivedAt = 918_251L;

        Provenance prov = Testimony.toldFactProvenance(tam, tellerLastSeenAt, receivedAt, 1);

        assertEquals(Provenance.Origin.TOLD, prov.origin);
        assertFalse(prov.origin.isDirect());
        assertEquals(tellerLastSeenAt, prov.observedAt, "observed_at is the TELLER's last-seen time (EH-5)");
        assertEquals(receivedAt, prov.receivedAt);
        assertEquals(tam, prov.source);
    }

    @Test
    void toldFactContentCarriesAssertion() {
        Place orchard = new Place("pl-51c", "the old orchard");
        Thing trees = Thing.of(null, "k:food-source", List.of("food_source"), "apple trees");

        Map<String, Object> content = Testimony.toldFactContent(orchard, trees, Testimony.Assertion.PRESENT);

        assertEquals("present", content.get("assertion"));
        assertEquals(orchard.toPayload(), content.get("about_place"));
        assertEquals(trees.toPayload(), content.get("thing"));
    }

    @Test
    void textIsReadOriginIndirectFromAnArtifact() {
        Provenance.Source board = Provenance.Source.artifact("th-12", "the job board");
        Provenance prov = Testimony.textProvenance(board, null, 918_260L);

        assertEquals(Provenance.Origin.READ, prov.origin);
        assertFalse(prov.origin.isDirect());
        assertEquals(board, prov.source);
        assertEquals(null, prov.observedAt, "an artifact's authorship time is typically unknown (§2.2)");
    }

    @Test
    void textContentAttributionIsClaimedNotVerified() {
        Map<String, Object> content = Testimony.textContent(
            "Build a shelter by the north wall. — the player", "the player");
        assertEquals("the player", content.get("attributed_to"));
        assertTrue(content.get("text").toString().contains("shelter"));
    }
}
