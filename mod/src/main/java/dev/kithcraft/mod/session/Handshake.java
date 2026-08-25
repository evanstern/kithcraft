package dev.kithcraft.mod.session;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Builds the {@code session_open} handshake (protocol §6.2, seam-wire-v0.md §1). The
 * capability manifest's real content — the vendor's actual declared verbs, salient kinds,
 * etc. — is Phase 3's (US2, T007) and must be L-7-safe (byte-identical for every body,
 * independent of world state). {@link #PLACEHOLDER_MANIFEST} is structurally correct
 * (the four-type floor, the core verbs, the protocol's fixed bearings/distance bands) but
 * is not the vendor's real capability declaration.
 */
public final class Handshake {
    private Handshake() {}

    /** PLACEHOLDER — Phase 3 (T007) replaces the content; the shape is what §6.2 requires. */
    public static final Map<String, Object> PLACEHOLDER_MANIFEST = buildPlaceholderManifest();

    private static Map<String, Object> buildPlaceholderManifest() {
        Map<String, Object> m = new LinkedHashMap<>();
        // §6.2's four-type floor: act_result, observation, sighting, speech.
        m.put("percept_types", List.of("sighting", "observation", "speech", "act_result"));
        m.put("origins", List.of("acted", "saw", "heard", "felt", "told", "read"));
        // §5.5's four core verbs; extended verbs are Phase 3's to declare.
        m.put("verbs", List.of(
            Map.of("verb", "go_to", "targets", List.of("place", "thing", "body")),
            Map.of("verb", "speak", "targets", List.of("body", "person", "none")),
            Map.of("verb", "attend", "targets", List.of("place", "none")),
            Map.of("verb", "wait", "targets", List.of("none"))));
        m.put("salient_kinds", List.of());
        m.put("bearings", List.of("ahead", "behind", "left", "right"));
        m.put("distance_bands", List.of("here", "near", "middling", "far"));
        return m;
    }

    /** Builds one body's {@code session_open} envelope. {@code seq} is 0: a body's
     * vendor-to-mind counter starts at 0 with its own {@code session_open} (§3.2). */
    public static Map<String, Object> buildSessionOpen(
            String session, String body, long worldTime, Continuity continuity) {
        Map<String, Object> payload = new LinkedHashMap<>();
        payload.put("time_unit", "second");
        payload.put("continuity", continuityPayload(continuity));
        payload.put("capabilities", PLACEHOLDER_MANIFEST);

        Map<String, Object> envelope = new LinkedHashMap<>();
        envelope.put("protocol", "0.1");
        envelope.put("message", "session_open");
        envelope.put("session", session);
        envelope.put("seq", 0L);
        envelope.put("body", body);
        envelope.put("world_time", worldTime);
        envelope.put("payload", payload);
        return envelope;
    }

    private static Map<String, Object> continuityPayload(Continuity c) {
        Map<String, Object> out = new LinkedHashMap<>();
        out.put("previous_session", c.previousSession());
        out.put("previous_close_world_time", c.previousCloseWorldTime());
        out.put("body_continuous", c.bodyContinuous());
        return out;
    }
}
