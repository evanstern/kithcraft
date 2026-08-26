package dev.kithcraft.mod.percept;

import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;
import java.util.Set;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T005 (card AC #5, §4.10): the delivery restriction — a change_report never reaches the body
 * that caused the change or a body that watched it happen. Three-body scenario: actor (caused
 * it), witness (saw it happen), absent (was elsewhere) — only the absent body may receive it. */
class ChangeReportsTest {

    @Test
    void onlyTheAbsentThirdPartyReceivesTheChangeReport() {
        Set<String> candidates = Set.of("b-actor", "b-witness", "b-absent");

        Set<String> recipients = ChangeReports.recipients(candidates, "b-actor", Set.of("b-witness"));

        assertEquals(Set.of("b-absent"), recipients);
        assertFalse(recipients.contains("b-actor"), "the actor already knows via its act_result (§5.4)");
        assertFalse(recipients.contains("b-witness"), "the witness already knows via its sighting");
        assertTrue(recipients.contains("b-absent"));
    }

    @Test
    void withNoActorOrWitnessesEveryCandidateReceivesIt() {
        Set<String> candidates = Set.of("b-1", "b-2");
        assertEquals(candidates, ChangeReports.recipients(candidates, null, null));
    }

    @Test
    void contentCarriesChangeAndThing() {
        Thing oak = Thing.of("th-77", "k:tree", List.of(), "the big oak");
        Map<String, Object> content = ChangeReports.content(ChangeReports.Change.GONE, oak);

        assertEquals("gone", content.get("change"));
        assertEquals(oak.toPayload(), content.get("thing"));
    }

    @Test
    void provenanceIsAlwaysSawDirectWithNullSource() {
        Provenance prov = ChangeReports.provenance(918300L, 918300L);
        assertEquals(Provenance.Origin.SAW, prov.origin);
        assertEquals(null, prov.source);
        assertTrue(prov.origin.isDirect());
    }
}
