package dev.kithcraft.mod.death;

import dev.kithcraft.mod.percept.Place;
import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

class GraveBoardEntryTest {

    @Test
    void contentIsATextPercept() {
        var entry = new GraveBoardEntry("Tam", new Place("pl-1", "Tam's grave"));
        Map<String, Object> content = entry.content();
        assertEquals("Tend Tam's grave.", content.get("text"));
        assertEquals(null, content.get("attributed_to"));
    }

    @Test
    void takeableOrIgnorable() {
        var entry = new GraveBoardEntry("Tam", new Place("pl-1", "Tam's grave"));
        assertFalse(entry.isTaken(), "ignorable by default");
        entry.take();
        assertTrue(entry.isTaken());
        entry.take(); // idempotent
        assertTrue(entry.isTaken());
    }
}
