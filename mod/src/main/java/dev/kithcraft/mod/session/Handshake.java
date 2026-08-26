package dev.kithcraft.mod.session;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Builds the {@code session_open} handshake (protocol §6.2, seam-wire-v0.md §1).
 *
 * <p>{@link #MANIFEST} is the vendor's real, static capability declaration (T007, US2). It
 * is a plain constant assembled with no world-query API in reach by construction, which is
 * what makes it L-7-safe: identical bytes for every body, in every world state, forever. It
 * declares only what this V1 skeleton can honestly promise — a declared capability is a
 * contract V2 must keep, so where honesty was in doubt the floor was declared and no more:
 *
 * <ul>
 *   <li>{@code percept_types} — the §6.2 floor ({@code act_result}, {@code observation},
 *       {@code sighting}, {@code speech}) plus the optional extras this vendor now composes
 *       ({@code self_state}, {@code sound}, {@code told_fact}, {@code text}, {@code
 *       change_report} — Phase 1/2, TASK-0012 V2). {@code sound} is declared because R-8
 *       verified a native hearing hook exists (specs/012-vendor-conformance/research/
 *       r8-hearing-hook.md); every extra here has a composing class in {@code percept/}.
 *   <li>{@code origins} — §2.7's full closed vocabulary (fixed by the protocol, not a vendor
 *       choice).
 *   <li>{@code verbs} — the §5.5 core four only ({@code go_to}, {@code speak}, {@code attend},
 *       {@code wait}); extended verbs (e.g. {@code carry}) are V2's to declare once built.
 *   <li>{@code salient_kinds} — a minimal, world-independent starter vocabulary the augmented
 *       villager's vanilla substrate already makes real (docs/design/entity-implementation-
 *       comparison.md R1.1/R1.2: beds, doors+bells in raid behavior, hostile targeting) plus
 *       the person kind every sighting needs: {@code k:person}, {@code k:sleeping-place},
 *       {@code k:door}, {@code k:hostile}.
 * </ul>
 */
public final class Handshake {
    private Handshake() {}

    /** The real V1 capability manifest (T007). See the class doc for what each field
     * declares and why. Never reads world state — this is what makes L-7 true. */
    public static final Map<String, Object> MANIFEST = buildManifest();

    private static Map<String, Object> buildManifest() {
        Map<String, Object> m = new LinkedHashMap<>();
        // §6.2's four-type floor plus the extras this vendor now emits (TASK-0012 Phase 1/2).
        m.put("percept_types", List.of("act_result", "observation", "sighting", "speech",
            "self_state", "sound", "told_fact", "text", "change_report"));
        // §2.7's full closed origin vocabulary.
        m.put("origins", List.of("acted", "saw", "heard", "felt", "told", "read"));
        // §5.5's four core verbs; extended verbs are V2's to declare.
        m.put("verbs", List.of(
            Map.of("verb", "go_to", "targets", List.of("place", "thing", "body")),
            Map.of("verb", "speak", "targets", List.of("body", "person", "none")),
            Map.of("verb", "attend", "targets", List.of("place", "none")),
            Map.of("verb", "wait", "targets", List.of("none"))));
        // A minimal, world-independent starter vocabulary (§3 AR-2/AR-3: opaque kind
        // tokens, roles + descriptor carry the meaning). Not derived from any world query.
        m.put("salient_kinds", List.of(
            Map.of("kind", "k:person", "roles", List.of(), "descriptor", "a person"),
            Map.of("kind", "k:sleeping-place", "roles", List.of("shelter", "sleeping_place"),
                "descriptor", "a bed"),
            Map.of("kind", "k:door", "roles", List.of("boundary"), "descriptor", "a door"),
            Map.of("kind", "k:hostile", "roles", List.of("danger"),
                "descriptor", "a hostile creature")));
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
        payload.put("capabilities", MANIFEST);

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
