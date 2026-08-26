package dev.kithcraft.mod.percept;

import dev.kithcraft.mod.wire.WireClient;

import java.io.IOException;
import java.util.LinkedHashMap;
import java.util.Locale;
import java.util.Map;
import java.util.Objects;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

/**
 * Composes and sends the percept envelope (§4.1) over the V1 transport: provenance stamped
 * at emission (by the caller, via {@link Provenance}), urgency bands (§2.8), and a per-body
 * monotonic {@code seq} assigned at ENQUEUE time. A shed percept still consumes its seq, so
 * a gap on the wire can only ever mean deliberate shedding (§4.11), never corruption.
 *
 * <p>An observation is composed and sent as one atomic call — {@link #compose} builds the
 * whole payload before anything reaches the sink, so "shed whole or delivered whole" (§4.11)
 * holds structurally; there is no partial-send path to guard against.
 *
 * <p>ponytail: this emitter's per-body {@code seq} starts fresh at 1 rather than continuing
 * {@link dev.kithcraft.mod.session.Handshake}'s {@code session_open} counter (which is
 * always 0) — the two meet only at a live session, which Phase 3/4 wires up. Ceiling: fine
 * for unit-scoped Phase 1: add a shared per-session sequence source when the emitter is
 * actually driven off a connected session.
 */
public final class PerceptEmitter {

    /** §2.8's urgency bands — the ONLY body-side signal. No salience/importance/weight/
     * memorability field is representable through this type, by construction. */
    public enum Urgency {
        BACKGROUND, NOTABLE, URGENT;

        public String wire() {
            return name().toLowerCase(Locale.ROOT);
        }
    }

    /** Where a composed envelope goes. {@link WireClient#send(Object)} satisfies this
     * directly; tests substitute a recording stub with no live socket. */
    @FunctionalInterface
    public interface Sink {
        void send(Object message) throws IOException;
    }

    private final Sink sink;
    private final String session;
    private final Map<String, AtomicLong> seqByBody = new ConcurrentHashMap<>();
    private final AtomicLong perceptCounter = new AtomicLong();

    public PerceptEmitter(Sink sink, String session) {
        this.sink = Objects.requireNonNull(sink, "sink");
        this.session = Objects.requireNonNull(session, "session");
    }

    public static PerceptEmitter overWireClient(WireClient client, String session) {
        return new PerceptEmitter(client::send, session);
    }

    /** Composes the percept envelope and assigns its {@code seq} — consumed here, at
     * enqueue, regardless of whether the caller goes on to send or shed it (§4.11). */
    public Map<String, Object> compose(String body, String perceptType, Urgency urgency,
            Provenance provenance, Place place, Map<String, Object> content, long worldTime) {
        Objects.requireNonNull(body, "body");
        Objects.requireNonNull(perceptType, "perceptType");
        Objects.requireNonNull(urgency, "urgency");
        Objects.requireNonNull(provenance, "provenance");
        Objects.requireNonNull(content, "content");

        Map<String, Object> percept = new LinkedHashMap<>();
        percept.put("percept_id", "p-" + perceptCounter.incrementAndGet());
        percept.put("percept_type", perceptType);
        percept.put("urgency", urgency.wire());
        percept.put("provenance", provenance.toPayload());
        percept.put("place", place == null ? null : place.toPayload());
        percept.put("content", content);

        long seq = seqByBody.computeIfAbsent(body, b -> new AtomicLong(0)).incrementAndGet();

        Map<String, Object> envelope = new LinkedHashMap<>();
        envelope.put("protocol", "0.1");
        envelope.put("message", "percept");
        envelope.put("session", session);
        envelope.put("seq", seq);
        envelope.put("body", body);
        envelope.put("world_time", worldTime);
        envelope.put("payload", percept);
        return envelope;
    }

    /** Composes and sends unconditionally. */
    public Map<String, Object> emit(String body, String perceptType, Urgency urgency,
            Provenance provenance, Place place, Map<String, Object> content, long worldTime)
            throws IOException {
        Map<String, Object> envelope = compose(body, perceptType, urgency, provenance, place, content, worldTime);
        sink.send(envelope);
        return envelope;
    }

    /** §4.11: under pressure a vendor MAY shed a {@code background} percept; it MUST NOT
     * shed {@code urgent}. @return the sent envelope, or {@code null} if shed. */
    public Map<String, Object> emitOrShed(String body, String perceptType, Urgency urgency,
            Provenance provenance, Place place, Map<String, Object> content, long worldTime,
            boolean shed) throws IOException {
        Map<String, Object> envelope = compose(body, perceptType, urgency, provenance, place, content, worldTime);
        if (shed) {
            if (urgency == Urgency.URGENT) {
                throw new IllegalArgumentException("urgent percepts MUST NOT be shed (§4.11)");
            }
            return null;
        }
        sink.send(envelope);
        return envelope;
    }
}
