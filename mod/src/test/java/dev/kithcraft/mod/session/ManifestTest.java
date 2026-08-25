package dev.kithcraft.mod.session;

import dev.kithcraft.mod.wire.CanonicalJson;
import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertSame;

/**
 * L-7 (protocol §6.2, card AC #3): the capability manifest describes the vendor, never the
 * world, so it MUST be byte-identical for every body in every world state.
 *
 * <p>The strongest form of this proof is structural, not incidental equality: {@code
 * Handshake}'s private manifest builder takes zero parameters, so world state is not even
 * an input it could branch on — {@link Handshake#MANIFEST} is one constant built once. This
 * test demonstrates that by building two full {@code session_open} envelopes for two
 * different bodies, with two different world times and two different continuity states
 * (simulating two different world states), and byte-comparing only the {@code capabilities}
 * object each one carries.
 */
class ManifestTest {

    @Test
    void manifestIsByteIdenticalAcrossBodiesAndWorldStates() {
        Map<String, Object> openA = Handshake.buildSessionOpen(
            "s-aaa", "b-aaa", 100L, Continuity.firstSession());
        Map<String, Object> openB = Handshake.buildSessionOpen(
            "s-bbb", "b-bbb", 918_233_000L,
            new Continuity("s-prior", 902_100L, true));

        Object capsA = payload(openA).get("capabilities");
        Object capsB = payload(openB).get("capabilities");

        assertArrayEquals(CanonicalJson.encode(capsA), CanonicalJson.encode(capsB),
            "capabilities must be byte-identical regardless of body or world state (L-7)");
    }

    @Test
    void manifestIsAStaticConstantNotABuilder() {
        // The manifest is assembled once, with no parameters, at class-init time. Two
        // different bodies necessarily observe the exact same object — the signature
        // itself makes L-7 structural rather than merely observed-to-hold.
        Map<String, Object> openA = Handshake.buildSessionOpen(
            "s-1", "b-1", 1L, Continuity.firstSession());
        Map<String, Object> openB = Handshake.buildSessionOpen(
            "s-2", "b-2", 2L, new Continuity("s-0", 0L, true));

        assertSame(Handshake.MANIFEST, payload(openA).get("capabilities"));
        assertSame(Handshake.MANIFEST, payload(openB).get("capabilities"));
    }

    @SuppressWarnings("unchecked")
    private static Map<String, Object> payload(Map<String, Object> envelope) {
        return (Map<String, Object>) envelope.get("payload");
    }
}
