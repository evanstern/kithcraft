package dev.kithcraft.mod.percept;

import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T003: provenance is stamped correctly per origin type, and EH-3's classifier matches the
 * protocol's closed vocabulary exactly. */
class ProvenanceTest {

    @Test
    void classifierMatchesTheProtocolTable() {
        for (Provenance.Origin o : new Provenance.Origin[] {
                Provenance.Origin.ACTED, Provenance.Origin.SAW,
                Provenance.Origin.HEARD, Provenance.Origin.FELT}) {
            assertTrue(o.isDirect(), o + " is direct (§2.7)");
        }
        for (Provenance.Origin o : new Provenance.Origin[] {Provenance.Origin.TOLD, Provenance.Origin.READ}) {
            assertFalse(o.isDirect(), o + " is not direct (§2.7)");
        }
    }

    @Test
    void directOriginRejectsASource() {
        // A direct origin routed through indirect()'s (source, hops) shape must still be
        // refused: the body is its own witness, so a direct percept MUST NOT carry a source.
        Provenance.Source someone = Provenance.Source.body("b-x", "someone");
        assertThrows(IllegalArgumentException.class,
            () -> Provenance.indirect(Provenance.Origin.SAW, someone, 1L, 1L, null));
    }

    @Test
    void indirectOriginRequiresASource() {
        assertThrows(NullPointerException.class,
            () -> Provenance.indirect(Provenance.Origin.TOLD, null, 1L, 2L, 0));
    }

    @Test
    void directPayloadHasNullSourceAndNoHops() {
        Provenance p = Provenance.direct(Provenance.Origin.SAW, 918233L, 918233L);
        Map<String, Object> payload = p.toPayload();
        assertEquals("saw", payload.get("origin"));
        assertNull(payload.get("source"));
        assertEquals(918233L, payload.get("observed_at"));
        assertEquals(918233L, payload.get("received_at"));
        assertFalse(payload.containsKey("hops"), "hops is informational-only and omitted when unset");
    }

    @Test
    void toldFactObservedAtIsTheTellerLastSeenTimeNotTheTellingTime() {
        Provenance.Source tam = Provenance.Source.body("b-tam", "Tam");
        // EH-5: observed_at = teller's last-seen time (902430), received_at = telling time (918251).
        Provenance p = Provenance.indirect(Provenance.Origin.TOLD, tam, 902430L, 918251L, 1);
        Map<String, Object> payload = p.toPayload();
        assertEquals("told", payload.get("origin"));
        assertEquals(902430L, payload.get("observed_at"));
        assertEquals(918251L, payload.get("received_at"));
        assertEquals(1L, payload.get("hops"));
        @SuppressWarnings("unchecked")
        Map<String, Object> source = (Map<String, Object>) payload.get("source");
        assertEquals("body", source.get("kind"));
        assertEquals("b-tam", source.get("body"));
    }

    @Test
    void observedAtNullMeansUnknown() {
        Provenance.Source artifact = Provenance.Source.artifact("th-12", "the job board");
        Provenance p = Provenance.indirect(Provenance.Origin.READ, artifact, null, 918260L, null);
        assertNull(p.toPayload().get("observed_at"));
    }

    @Test
    void sourceIsTheImmediateTellerNeverThePromotedOrigin() {
        // A twice-relayed fact still carries origin "told" with hops growing — relaying
        // never launders secondhand into direct (§2.6).
        Provenance.Source relay = Provenance.Source.body("b-mid", "the middle teller");
        Provenance p = Provenance.indirect(Provenance.Origin.TOLD, relay, 900000L, 918000L, 2);
        assertEquals("told", p.toPayload().get("origin"));
        assertEquals(2L, p.toPayload().get("hops"));
        assertFalse(p.origin.isDirect());
    }
}
