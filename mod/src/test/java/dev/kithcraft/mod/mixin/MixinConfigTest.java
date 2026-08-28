package dev.kithcraft.mod.mixin;

import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * T005 (FR-004, card AC #3) + TASK-0019 T004/T005 (FR-002/FR-003, card ACs #2/#3): the Mixin
 * surface is bounded — decision-0002's ~4 targeted-injection ceiling, raised from V3's 3-slot
 * bound once siege-suppression and conversion-cancel landed
 * (specs/019-death-remains/research/death-26.2.md STOP/GO budget check) — and enumerated, each
 * with a stated purpose. Reads {@code kithcraft.mixins.json} directly (not via Sponge Mixin's
 * own loader) so this test needs no live Minecraft classes.
 */
class MixinConfigTest {

    @Test
    void configListsAtMostFourMixinsEachWithAStatedPurpose() throws IOException {
        JsonObject config;
        try (InputStream in = getClass().getClassLoader().getResourceAsStream("kithcraft.mixins.json")) {
            assertNotNull(in, "kithcraft.mixins.json must be on the classpath");
            config = JsonParser.parseReader(new InputStreamReader(in, StandardCharsets.UTF_8)).getAsJsonObject();
        }

        var mixins = config.getAsJsonArray("mixins");
        assertFalse(mixins.isEmpty(), "at least one override is expected (breeding/gossip suppression)");
        assertTrue(mixins.size() <= 4, "decision-0002/card AC #3: at most four targeted overrides, got " + mixins.size());

        var purposes = config.getAsJsonObject("kithcraftPurposes");
        assertNotNull(purposes, "each mixin must be named with its purpose (kithcraftPurposes)");
        for (var element : mixins) {
            String name = element.getAsString();
            assertTrue(purposes.has(name), "no purpose recorded for " + name);
            assertFalse(purposes.get(name).getAsString().isBlank(), "empty purpose for " + name);
        }
    }
}
