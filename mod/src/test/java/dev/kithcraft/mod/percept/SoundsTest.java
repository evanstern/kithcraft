package dev.kithcraft.mod.percept;

import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** R-8 (§4.4): sound content shape and heard/direct provenance with a null source. */
class SoundsTest {

    @Test
    void contentCarriesKindBearingDistanceDescriptor() {
        Map<String, Object> content = Sounds.content("k:snarl", "behind", "near", "something snarling in the dark");
        assertEquals("k:snarl", content.get("sound_kind"));
        assertEquals("behind", content.get("bearing"));
        assertEquals("near", content.get("distance"));
        assertTrue(content.get("descriptor").toString().contains("snarling"));
    }

    @Test
    void contentHasNoFieldNamingTheCauser() {
        Map<String, Object> content = Sounds.content("k:snarl", "behind", "near", "something snarling in the dark");
        assertFalse(content.containsKey("thing_id"));
        assertFalse(content.containsKey("thing"));
        assertFalse(content.containsKey("kind") && !content.containsKey("sound_kind"));
    }

    @Test
    void provenanceIsHeardDirectWithNullSource() {
        Provenance prov = Sounds.provenance(918_244L, 918_244L);
        assertEquals(Provenance.Origin.HEARD, prov.origin);
        assertTrue(prov.origin.isDirect());
        assertEquals(null, prov.source);
    }
}
