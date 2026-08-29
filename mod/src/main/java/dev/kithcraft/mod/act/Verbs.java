package dev.kithcraft.mod.act;

import dev.kithcraft.mod.percept.Place;

import java.util.Map;
import java.util.Set;

/**
 * §5.5's core four verbs, executed against a world-facing seam so unit tests script them
 * deterministically without a live server ({@code T011} wires the real dev-server
 * implementation). Movement/look/say are plain vanilla {@code Brain<E>}/entity API per
 * decision-0002 — this class adds no Mixin.
 */
public final class Verbs {
    private Verbs() {}

    /** The declared verb set (mirrors {@code Handshake.MANIFEST}'s {@code verbs} list) —
     * the single source of truth {@link IntentHandler} checks an incoming intent's {@code
     * verb} against for §5.3's {@code unknown_verb} refusal (V-4). {@code claim} (TASK-0020
     * T004) is the core four plus one manifest-declared extended verb — the same "extras"
     * mechanism {@code Handshake.MANIFEST}'s {@code percept_types} already uses, applied to
     * verbs (plan.md design decision 3): the core floor has nothing shaped like "take this
     * posting," so it rides §5.5's declared-extras half rather than a protocol change. */
    public static final Set<String> DECLARED = Set.of("go_to", "speak", "attend", "wait", "claim");

    /** One tick's worth of {@code go_to} progress toward a resolved place. {@link
     * IntentHandler} calls {@link Actuator#stepWalk} once per tick until {@link #done()}. */
    public enum WalkStatus {
        IN_PROGRESS, ARRIVED,
        /** The referent still exists but no path could be found (§5.4's {@code
         * unreachable}). */
        UNREACHABLE,
        /** §5.6's one sanctioned discovery channel: the referent the walk was headed toward
         * no longer exists. Distinct from {@link #UNREACHABLE} — the walk itself succeeded;
         * what it found there did not. */
        TARGET_GONE;

        public boolean done() {
            return this != IN_PROGRESS;
        }
    }

    /** The world-facing seam every verb drives. Real wiring (T011) targets vanilla {@code
     * PathNavigation}/{@code LookControl}/a chat-adjacent call, no Mixin (decision-0002);
     * tests substitute a scripted stub — the same pattern as {@link
     * dev.kithcraft.mod.percept.Sightings.HostileSource} and {@link
     * dev.kithcraft.mod.percept.PerceptEmitter.Sink}. */
    public interface Actuator {
        /** Starts or continues walking {@code body} toward {@code place}, one tick's worth
         * of progress per call. */
        WalkStatus stepWalk(String body, Place place);

        /** Fires immediately — {@code speak} has no multi-tick progress (§9.2's Hearth
         * sketch: "append the utterance ... emit speech percepts", one beat). {@code target}
         * is the intent's raw target object, opaque to this seam. */
        void speak(String body, Map<String, Object> target, String text);

        /** Looks deliberately at {@code place} (or the body's own facing, for a {@code none}
         * target); fires immediately. */
        void attend(String body, Place place);
    }
}
