package dev.kithcraft.mod.act;

import dev.kithcraft.mod.percept.PerceptEmitter;
import dev.kithcraft.mod.percept.Place;
import dev.kithcraft.mod.percept.Provenance;

import java.io.IOException;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.concurrent.atomic.AtomicLong;

/**
 * §5's act surface: {@code intent} decode → {@code intent_ack} (receipt/admissibility only,
 * §5.3) → exactly one {@code act_result} percept per accepted intent (§5.4), plus {@code
 * cancel} (§5.7). One instance serves one body.
 *
 * <p>Only {@code go_to} is multi-tick: it registers a pending walk and {@link #tick} drives
 * it to completion across calls. {@code speak}/{@code attend}/{@code wait} complete inline
 * within {@link #handle} — the protocol gives no duration field for {@code wait} (§5.5's
 * table has none), so a {@code wait} intent is one beat's worth of "did nothing, on purpose"
 * (ponytail: add a duration convention if/when the protocol grows one; a mind re-issues
 * {@code wait} to keep waiting, same as Hearth's own worked sketch, §9.2).
 *
 * <p>TASK-0020 T004 (US2): {@code claim} is a fifth verb, also inline-completing, resolved
 * through {@link ClaimRegistry} rather than {@link Verbs.Actuator} — claiming is board-side
 * bookkeeping, not a world-facing act. It is declared via §5.5's extended-verb mechanism
 * (manifest-declared, {@link dev.kithcraft.mod.session.Handshake}), not the core floor: none
 * of {@code go_to}/{@code speak}/{@code attend}/{@code wait} fits "take this posting."
 */
public final class IntentHandler {

    /** One accepted, not-yet-resolved {@code go_to}. Everything needed to finish it and
     * compose its eventual {@code act_result} without re-consulting anything else. */
    private record Pending(String verb, Place place, String reason, Long notAfter) {}

    private final PerceptEmitter emitter;
    private final TargetResolution.Ground ground;
    private final Verbs.Actuator actuator;
    private final ClaimRegistry claims;
    private final String session;
    private final String body;
    private final Map<String, Pending> pendingWalks = new LinkedHashMap<>();
    // ponytail: a separate counter from PerceptEmitter's per-body percept seq, per that
    // class's own standing note that the two converge only once a live session drives both.
    // Ceiling: fine while nothing consumes intent_ack's seq for ordering; unify when a real
    // session wires percept + ack onto one vendor->mind stream.
    private final AtomicLong ackSeq = new AtomicLong();

    public IntentHandler(PerceptEmitter emitter, TargetResolution.Ground ground,
            Verbs.Actuator actuator, ClaimRegistry claims, String session, String body) {
        this.emitter = Objects.requireNonNull(emitter, "emitter");
        this.ground = Objects.requireNonNull(ground, "ground");
        this.actuator = Objects.requireNonNull(actuator, "actuator");
        this.claims = Objects.requireNonNull(claims, "claims");
        this.session = Objects.requireNonNull(session, "session");
        this.body = Objects.requireNonNull(body, "body");
    }

    /** Decodes and admits (or refuses) one {@code intent} payload, returning the {@code
     * intent_ack} envelope. An accepted {@code go_to} is registered and given its first
     * step immediately; other accepted verbs complete inline, each emitting exactly one
     * {@code act_result} (card AC #6) before this method returns. */
    @SuppressWarnings("unchecked")
    public Map<String, Object> handle(Map<String, Object> intent, long worldTime) throws IOException {
        Objects.requireNonNull(intent, "intent");
        String intentId = (String) intent.get("intent_id");
        String verb = (String) intent.get("verb");
        if (intentId == null || verb == null) {
            return ack(intentId, false, "malformed");
        }
        if (!Verbs.DECLARED.contains(verb)) {
            return ack(intentId, false, "unknown_verb");
        }

        Map<String, Object> targetMap = (Map<String, Object>) intent.get("target");
        TargetResolution.Result resolved = TargetResolution.resolve(ground, targetMap);
        if (resolved.outcome() == TargetResolution.Outcome.UNKNOWN_TARGET) {
            return ack(intentId, false, "unknown_target");
        }

        String reason = (String) intent.get("reason");
        String supersedes = (String) intent.get("supersedes");
        Long notAfter = intent.get("not_after") == null ? null : ((Number) intent.get("not_after")).longValue();
        if (supersedes != null) {
            supersede(supersedes, worldTime);
        }

        Map<String, Object> ack = ack(intentId, true, null);
        switch (verb) {
            case "go_to" -> {
                pendingWalks.put(intentId, new Pending(verb, resolved.place(), reason, notAfter));
                advanceWalk(intentId, worldTime);
            }
            case "speak" -> {
                Map<String, Object> content = (Map<String, Object>) intent.get("content");
                String text = content == null ? null : (String) content.get("text");
                actuator.speak(body, targetMap, text);
                emitResult(intentId, verb, resolved.place(), "completed", null, reason,
                    text == null ? "said nothing" : "said: " + text, worldTime);
            }
            case "attend" -> {
                actuator.attend(body, resolved.place());
                emitResult(intentId, verb, resolved.place(), "completed", null, reason,
                    "looked around", worldTime);
            }
            case "wait" -> emitResult(intentId, verb, null, "completed", null, reason, "waited", worldTime);
            case "claim" -> {
                // T004 (AR-4): the mind named a token in a percept it was shown (the board's
                // own thing token, riding provenance.source on every read — Testimony's §4.7
                // shape); this is the engine resolving it. No new target field: "thing" is
                // the SAME shape go_to/speak already parse (TargetResolution).
                String thingToken = targetMap == null ? null : (String) targetMap.get("thing_id");
                switch (claims.claim(body, thingToken)) {
                    case CLAIMED -> emitResult(intentId, verb, null, "completed", null, reason,
                        "claimed it", worldTime);
                    case ALREADY_CLAIMED -> emitResult(intentId, verb, null, "failed", "blocked",
                        reason, "someone else already claimed it", worldTime);
                    case UNKNOWN_POSTING -> emitResult(intentId, verb, null, "failed", "not_capable",
                        reason, "that isn't a postable thing", worldTime);
                }
            }
            default -> throw new AssertionError("unreachable: " + verb + " is in Verbs.DECLARED");
        }
        return ack;
    }

    /** Advances every pending {@code go_to}, abandoning any whose {@code not_after} has
     * passed with {@code expired} (§5.2). Call once per server tick. */
    public void tick(long worldTime) throws IOException {
        for (String intentId : List.copyOf(pendingWalks.keySet())) {
            Pending p = pendingWalks.get(intentId);
            if (p == null) {
                continue; // superseded/cancelled since the snapshot above
            }
            if (p.notAfter() != null && worldTime > p.notAfter()) {
                pendingWalks.remove(intentId);
                emitResult(intentId, p.verb(), null, "expired", null, p.reason(),
                    "gave up: past not_after", worldTime);
                continue;
            }
            advanceWalk(intentId, worldTime);
        }
    }

    /** §5.7 {@code cancel}: requests abandonment of a pending intent. A no-op if the intent
     * is not (or no longer) pending — "cancellation is a request, not a guarantee" (§5.7). */
    public void cancel(String intentId, long worldTime) throws IOException {
        Pending p = pendingWalks.remove(intentId);
        if (p != null) {
            emitResult(intentId, p.verb(), null, "superseded", null, p.reason(),
                "cancelled before completion", worldTime);
        }
    }

    private void supersede(String intentId, long worldTime) throws IOException {
        cancel(intentId, worldTime);
    }

    private void advanceWalk(String intentId, long worldTime) throws IOException {
        Pending p = pendingWalks.get(intentId);
        if (p == null) {
            return;
        }
        if (p.place() == null) {
            // Issued token, but nothing has ever recorded a place for it — there is nowhere
            // to path to. Not unknown_target (the token IS issued, §5.3); reported as
            // unreachable rather than target_gone, which is reserved for a referent found
            // gone by actually trying to reach it (§5.6).
            pendingWalks.remove(intentId);
            emitResult(intentId, p.verb(), null, "failed", "unreachable", p.reason(),
                "never saw where that was", worldTime);
            return;
        }
        Verbs.WalkStatus status = actuator.stepWalk(body, p.place());
        if (!status.done()) {
            return;
        }
        pendingWalks.remove(intentId);
        switch (status) {
            case ARRIVED -> emitResult(intentId, p.verb(), p.place(), "completed", null, p.reason(),
                "walked to " + descriptor(p.place()), worldTime);
            case UNREACHABLE -> emitResult(intentId, p.verb(), null, "failed", "unreachable", p.reason(),
                "could not find a way there", worldTime);
            case TARGET_GONE -> emitResult(intentId, p.verb(), p.place(), "failed", "target_gone", p.reason(),
                "got there and found nothing", worldTime);
            case IN_PROGRESS -> throw new AssertionError("unreachable: status.done() was true");
        }
    }

    private static String descriptor(Place place) {
        return place.descriptor() == null ? place.place() : place.descriptor();
    }

    private Map<String, Object> ack(String intentId, boolean accepted, String reasonCode) {
        Map<String, Object> payload = new LinkedHashMap<>();
        payload.put("intent_id", intentId);
        payload.put("accepted", accepted);
        payload.put("reason_code", reasonCode);

        Map<String, Object> envelope = new LinkedHashMap<>();
        envelope.put("protocol", "0.1");
        envelope.put("message", "intent_ack");
        envelope.put("session", session);
        envelope.put("seq", ackSeq.incrementAndGet());
        envelope.put("body", body);
        envelope.put("payload", payload);
        return envelope;
    }

    /** Composes and sends the one {@code act_result} percept an accepted intent owes
     * (§5.4, card AC #6). {@code place} is the arrival/attended place when the verb
     * resolved one; {@code null} for a verb with no place target (speak/wait) — the body's
     * own live position is T011 dev-server wiring's job (a plain, no-Mixin entity-position
     * read), not something this unit-tested state machine can supply. */
    private void emitResult(String intentId, String verb, Place place, String outcome,
            String reasonCode, String reason, String detail, long worldTime) throws IOException {
        Map<String, Object> content = new LinkedHashMap<>();
        content.put("intent_id", intentId);
        content.put("verb", verb);
        content.put("outcome", outcome);
        content.put("reason_code", reasonCode);
        content.put("reason", reason);
        content.put("detail", detail);

        Provenance provenance = Provenance.direct(Provenance.Origin.ACTED, worldTime, worldTime);
        emitter.emit(body, "act_result", PerceptEmitter.Urgency.NOTABLE, provenance, place, content, worldTime);
    }
}
