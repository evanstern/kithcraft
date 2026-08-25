package dev.kithcraft.mod;

import dev.kithcraft.mod.act.IntentHandler;
import dev.kithcraft.mod.act.TargetResolution;
import dev.kithcraft.mod.act.Verbs;
import dev.kithcraft.mod.percept.ChangeReports;
import dev.kithcraft.mod.percept.PerceptEmitter;
import dev.kithcraft.mod.percept.Place;
import dev.kithcraft.mod.percept.Provenance;
import dev.kithcraft.mod.percept.SelfState;
import dev.kithcraft.mod.percept.Sightings;
import dev.kithcraft.mod.percept.Sounds;
import dev.kithcraft.mod.percept.Testimony;
import dev.kithcraft.mod.percept.Thing;
import dev.kithcraft.mod.session.Continuity;
import dev.kithcraft.mod.session.Handshake;
import org.junit.jupiter.api.Test;

import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.regex.Pattern;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * T010 (card AC #9): body-protocol-v0.md §12's six leak passes, run over payloads composed
 * through the real Phase 1-3 machinery (the emitters/handlers themselves, not hand-typed
 * fixtures) — session_open, every percept type, every intent_ack outcome, every act_result
 * outcome across all four verbs. §12.2's own findings (L-1..L-10) are the allowlist and the
 * regression set this test exists to keep true.
 */
class LeakPassTest {

    // ---- captured payloads: composed through the real classes, not hand-typed fixtures ----

    private static List<Object> capture() throws Exception {
        List<Object> captured = new ArrayList<>();

        // P-5 fixture pair: two session_opens for different bodies/sessions/world times —
        // proves the manifest itself is invariant (L-7) rather than merely checking one.
        captured.add(Handshake.buildSessionOpen("s-1", "b-1", 100L, Continuity.firstSession()));
        captured.add(Handshake.buildSessionOpen("s-2", "b-2", 999_999L, Continuity.firstSession()));

        List<Object> sent = new ArrayList<>();
        PerceptEmitter emitter = new PerceptEmitter(sent::add, "s-1");

        Thing bed = Thing.of(null, "k:sleeping-place", List.of("shelter"), "a bed");
        Thing person = new Thing("th-9", "k:person", List.of(), "Tam", "b-tam", 1);
        Provenance saw = Provenance.direct(Provenance.Origin.SAW, 100L, 100L);
        Provenance felt = SelfState.provenance(100L, 100L);
        Provenance heard = Sounds.provenance(100L, 100L);
        Provenance told = Testimony.toldFactProvenance(
            Provenance.Source.body("b-tam", "Tam"), 90L, 100L, 1);
        Provenance read = Testimony.textProvenance(
            Provenance.Source.artifact("th-2", "a sign"), null, 100L);

        emitter.emit("b-eda", "sighting", PerceptEmitter.Urgency.NOTABLE, saw,
            new Place("pl-1", "the well"), Sightings.sightingContent(person, "near", "drawing water"), 100L);
        emitter.emit("b-eda", "observation", PerceptEmitter.Urgency.BACKGROUND, saw,
            new Place("pl-1", "the well"),
            Sightings.observationContent("near", List.of("k:sleeping-place", "k:person"), List.of(bed)), 100L);
        emitter.emit("b-eda", "self_state", PerceptEmitter.Urgency.BACKGROUND, felt, null,
            SelfState.content(SelfState.HUNGER, SelfState.Level.MILD, SelfState.Trend.WORSENING), 100L);
        emitter.emit("b-eda", "sound", PerceptEmitter.Urgency.NOTABLE, heard, null,
            Sounds.content("footsteps", "behind", "near", "something moving"), 100L);
        emitter.emit("b-eda", "told_fact", PerceptEmitter.Urgency.NOTABLE, told,
            new Place("pl-1", "the well"),
            Testimony.toldFactContent(new Place("pl-2", "the market"), bed, Testimony.Assertion.PRESENT), 100L);
        emitter.emit("b-eda", "text", PerceptEmitter.Urgency.BACKGROUND, read, null,
            Testimony.textContent("meet at dusk", "Tam"), 100L);
        emitter.emit("b-other", "change_report", PerceptEmitter.Urgency.BACKGROUND, saw,
            new Place("pl-1", "the well"), ChangeReports.content(ChangeReports.Change.GONE, bed), 100L);
        captured.addAll(sent);

        // intent_ack + act_result across every verb and every refusal/outcome shape, via the
        // real IntentHandler state machine (not hand-typed envelopes).
        captured.addAll(driveIntentHandler());

        return captured;
    }

    private static List<Object> driveIntentHandler() throws Exception {
        List<Object> sent = new ArrayList<>();
        Place place = new Place("pl-1", "the well");
        TargetResolution.Ground ground = new TargetResolution.Ground() {
            @Override public boolean isIssued(String token) { return "pl-1".equals(token); }
            @Override public Optional<Place> placeFor(String token) {
                return "pl-1".equals(token) ? Optional.of(place) : Optional.empty();
            }
        };
        Deque<Verbs.WalkStatus> script = new ArrayDeque<>(List.of(
            Verbs.WalkStatus.ARRIVED, Verbs.WalkStatus.UNREACHABLE, Verbs.WalkStatus.IN_PROGRESS,
            Verbs.WalkStatus.TARGET_GONE));
        Verbs.Actuator actuator = new Verbs.Actuator() {
            @Override public Verbs.WalkStatus stepWalk(String body, Place p) { return script.poll(); }
            @Override public void speak(String body, Map<String, Object> target, String text) {}
            @Override public void attend(String body, Place p) {}
        };
        IntentHandler handler = new IntentHandler(new PerceptEmitter(sent::add, "s-1"), ground, actuator, "s-1", "b-eda");

        List<Object> envelopes = new ArrayList<>();
        envelopes.add(handler.handle(Map.of("intent_id", "i-1", "verb", "go_to",
            "target", Map.of("type", "place", "place", "pl-1")), 100L)); // arrives
        envelopes.add(handler.handle(Map.of("intent_id", "i-2", "verb", "speak",
            "target", Map.of("type", "none"), "content", Map.of("text", "hello")), 101L));
        envelopes.add(handler.handle(Map.of("intent_id", "i-3", "verb", "attend",
            "target", Map.of("type", "place", "place", "pl-1")), 102L));
        envelopes.add(handler.handle(Map.of("intent_id", "i-4", "verb", "wait",
            "target", Map.of("type", "none")), 103L));
        envelopes.add(handler.handle(Map.of("intent_id", "i-5", "verb", "flurbosplat"), 104L)); // unknown_verb
        envelopes.add(handler.handle(Map.of("intent_id", "i-6", "verb", "go_to",
            "target", Map.of("type", "body", "body", "b-ghost")), 105L)); // unknown_target

        Map<String, Object> unreachable = handler.handle(Map.of("intent_id", "i-7", "verb", "go_to",
            "target", Map.of("type", "place", "place", "pl-1")), 106L);
        envelopes.add(unreachable);
        Map<String, Object> gone = handler.handle(Map.of("intent_id", "i-8", "verb", "go_to",
            "target", Map.of("type", "place", "place", "pl-1")), 107L);
        envelopes.add(gone);
        handler.tick(108L); // resolves i-8's target_gone

        envelopes.addAll(sent);
        return envelopes;
    }

    // ---------------------------------------------------------------- P-1

    /** Every engine-native term — type names, registry identifiers, class names — in every
     * captured string. */
    private static final Pattern ENGINE_NATIVE = Pattern.compile(
        "(?i)\\bminecraft\\b|net\\.minecraft|fabricmc|\\bmixin\\b|\\bmob\\b(?!ile)");

    @Test
    void p1NoEngineNativeTermAnywhere() throws Exception {
        for (Object payload : capture()) {
            walkStrings(payload, s -> assertFalse(ENGINE_NATIVE.matcher(s).find(),
                "P-1: engine-native term leaked in \"" + s + "\""));
        }
    }

    // ---------------------------------------------------------------- P-2

    private static final Set<String> COORDINATE_KEYS =
        Set.of("x", "y", "z", "pos", "position", "blockpos", "block_pos");
    private static final Set<String> FORBIDDEN_SALIENCE_KEYS =
        Set.of("salience", "importance", "weight", "memorability");

    /** Every field the mind branches on is opaque, never a raw coordinate or a namespaced
     * registry id (AR-1/AR-2/AR-4). */
    @Test
    void p2NoStructuralLeakInAnyBranchableField() throws Exception {
        for (Object payload : capture()) {
            walkMaps(payload, (key, value) -> {
                assertFalse(COORDINATE_KEYS.contains(key.toLowerCase(Locale.ROOT)),
                    "P-2: coordinate-shaped field \"" + key + "\" (AR-4)");
                assertFalse(FORBIDDEN_SALIENCE_KEYS.contains(key),
                    "P-2/AC#4: forbidden salience-shaped field \"" + key + "\"");
                if ("kind".equals(key) && value instanceof String kind) {
                    // Two legitimate "kind" namespaces share this key name: a Thing's opaque
                    // k:-prefixed token (AR-1/AR-2), and Provenance.Source's own closed
                    // vocabulary (body/person/artifact/unknown, §2.6) — neither is a leak.
                    assertTrue(!kind.contains(":") || kind.startsWith("k:"),
                        "P-2: kind token \"" + kind + "\" is not opaque (AR-1/AR-2)");
                }
            });
        }
    }

    // ---------------------------------------------------------------- P-3

    /** Free-text fields exist and are exactly the vendor-declared prose-only set — nothing
     * else carries arbitrary text that could smuggle structure past grep (§12.1's own
     * warning: "free text is where structured leaks hide"). */
    private static final Set<String> PROSE_ONLY_KEYS =
        Set.of("descriptor", "detail", "doing", "text", "reason", "attributed_to");

    @Test
    void p3FreeTextFieldsAreOnlyTheDeclaredProseOnlySet() throws Exception {
        Set<String> seen = new java.util.LinkedHashSet<>();
        for (Object payload : capture()) {
            walkMaps(payload, (key, value) -> {
                if (value instanceof String s && s.contains(" ")) { // a sentence-shaped value
                    seen.add(key);
                }
            });
        }
        seen.removeAll(PROSE_ONLY_KEYS);
        assertTrue(seen.isEmpty(), "P-3: sentence-shaped value(s) outside the declared prose-only fields: " + seen);
    }

    // ---------------------------------------------------------------- P-4

    /** {@code intent_ack} acknowledges receipt only — no field beyond {@code intent_id},
     * {@code accepted}, {@code reason_code} — and never answers "does it still exist"
     * synchronously (L-6: {@code unknown_target} is refused only for an unissued token). */
    @Test
    void p4IntentAckCarriesReceiptOnlyNoExistenceOracle() throws Exception {
        for (Object payload : capture()) {
            if (!(payload instanceof Map<?, ?> m) || !"intent_ack".equals(m.get("message"))) {
                continue;
            }
            @SuppressWarnings("unchecked")
            Map<String, Object> ack = (Map<String, Object>) ((Map<String, Object>) m).get("payload");
            assertEquals(Set.of("intent_id", "accepted", "reason_code"), ack.keySet(),
                "P-4: intent_ack must carry receipt only, no extra world-state field");
        }
        // i-6 targets an unissued token -> refused unknown_target; i-7/i-8 target the known,
        // issued "pl-1" -> accepted, and only fail (unreachable/target_gone) after the walk.
        // No refusal in this capture answers "gone?" synchronously — the existence-oracle
        // hole L-6 closed.
    }

    // ---------------------------------------------------------------- P-5

    /** {@code session_open.capabilities} describes the vendor, not the world (L-7): byte-
     * identical across different bodies, sessions, and world times. */
    @Test
    @SuppressWarnings("unchecked")
    void p5ManifestIsVendorInvariantNotWorldDescribing() throws Exception {
        List<Object> captured = capture();
        Map<String, Object> a = (Map<String, Object>) captured.get(0);
        Map<String, Object> b = (Map<String, Object>) captured.get(1);
        Map<String, Object> capsA = (Map<String, Object>) ((Map<String, Object>) a.get("payload")).get("capabilities");
        Map<String, Object> capsB = (Map<String, Object>) ((Map<String, Object>) b.get("payload")).get("capabilities");
        assertEquals(capsA, capsB, "P-5: capabilities must be byte-identical across bodies/sessions/world_time (L-7)");
    }

    // ---------------------------------------------------------------- P-6

    private static final Set<String> DISTANCE_BANDS = Set.of("here", "near", "middling", "far");
    private static final Set<String> BEARINGS = Set.of("ahead", "behind", "left", "right");

    /** Every distance/extent/bearing field is a coarse band, never a raw unit — no
     * arithmetic affordance for reconstructing geometry (AR-4). */
    @Test
    void p6DistanceBearingExtentAreCoarseBandsOnly() throws Exception {
        for (Object payload : capture()) {
            walkMaps(payload, (key, value) -> {
                if (value == null) {
                    return;
                }
                if (("distance".equals(key) || "extent".equals(key)) ) {
                    assertTrue(DISTANCE_BANDS.contains(value), "P-6: \"" + key + "\" not a coarse band: " + value);
                }
                if ("bearing".equals(key)) {
                    assertTrue(BEARINGS.contains(value), "P-6: bearing not a coarse band: " + value);
                }
            });
        }
    }

    // ---------------------------------------------------------------- walkers

    @SuppressWarnings("unchecked")
    private static void walkStrings(Object v, java.util.function.Consumer<String> check) {
        if (v instanceof String s) {
            check.accept(s);
        } else if (v instanceof Map<?, ?> m) {
            m.values().forEach(val -> walkStrings(val, check));
        } else if (v instanceof List<?> l) {
            l.forEach(val -> walkStrings(val, check));
        }
    }

    @FunctionalInterface
    private interface EntryCheck {
        void accept(String key, Object value);
    }

    @SuppressWarnings("unchecked")
    private static void walkMaps(Object v, EntryCheck check) {
        if (v instanceof Map<?, ?> m) {
            for (Map.Entry<?, ?> e : m.entrySet()) {
                check.accept((String) e.getKey(), e.getValue());
                walkMaps(e.getValue(), check);
            }
        } else if (v instanceof List<?> l) {
            l.forEach(item -> walkMaps(item, check));
        }
    }
}
