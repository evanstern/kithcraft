package dev.kithcraft.mod.percept;

import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T002: self_state content shape and its always-felt, always-direct provenance. */
class SelfStateTest {

    @Test
    void contentCarriesConditionLevelAndTrendAsBands() {
        Map<String, Object> content = SelfState.content(SelfState.COLD, SelfState.Level.SEVERE, SelfState.Trend.WORSENING);
        assertEquals("cold", content.get("condition"));
        assertEquals("severe", content.get("level"));
        assertEquals("worsening", content.get("trend"));
    }

    @Test
    void trendMayBeNull() {
        Map<String, Object> content = SelfState.content(SelfState.FATIGUE, SelfState.Level.MILD, null);
        assertNull(content.get("trend"));
    }

    @Test
    void provenanceIsAlwaysFeltAndDirect() {
        Provenance p = SelfState.provenance(918270L, 918270L);
        assertEquals("felt", p.toPayload().get("origin"));
        assertTrue(p.origin.isDirect());
        assertNull(p.toPayload().get("source"), "the body is the only witness to its own condition (§2.7)");
    }
}
