package dev.kithcraft.mod.wire;

import org.junit.jupiter.api.DynamicTest;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestFactory;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.channels.Channels;
import java.nio.channels.ReadableByteChannel;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HexFormat;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeMap;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.fail;

/**
 * The 17 golden vectors (seam/vectors/) run against THIS mod's own codec —
 * {@link FrameCodec} + {@link CanonicalJson} — mirroring the obligations
 * seam/java-roundtrip/RoundTrip.java proved against a throwaway harness (TASK-0007).
 * Census (exactly 17, missing or extra both fail); per vector: decode, structural
 * equality with {@code decoded}, canonical re-encode byte-equality with {@code frame_hex},
 * and the pinned validation behavior ({@code roundtrip}/{@code decode_ok}/{@code refuse}).
 */
class VectorSuiteTest {

    /** contracts/vectors.md's closed list — extras are scope creep, gaps are holes. */
    private static final List<String> CENSUS = List.of(
        "percept_sighting", "percept_observation", "percept_sound", "percept_speech",
        "percept_told_fact", "percept_text", "percept_act_result", "percept_self_state",
        "percept_change_report",
        "intent", "intent_ack", "cancel",
        "session_open", "session_close",
        "err_missing_provenance", "err_unknown_origin", "intent_ack_refused");

    /** Required payload members per message type (✔ in the protocol's tables). */
    private static final Map<String, List<String>> REQUIRED_PAYLOAD = Map.of(
        "percept", List.of("percept_id", "percept_type", "urgency", "provenance", "place", "content"),
        "intent", List.of("intent_id", "verb", "target", "reason"),
        "intent_ack", List.of("intent_id", "accepted", "reason_code"),
        "cancel", List.of("intent_id"),
        "session_open", List.of("time_unit", "capabilities"),
        "session_close", List.of("reason", "detail"));

    /** Resolves seam/vectors relative to the test working directory, robustly: prefer the
     * path Gradle passes explicitly, else walk up from the working directory. */
    private static Path vectorsDir() {
        String configured = System.getProperty("kithcraft.seamVectorsDir");
        if (configured != null && Files.isDirectory(Path.of(configured))) {
            return Path.of(configured);
        }
        for (Path dir = Path.of("").toAbsolutePath(); dir != null; dir = dir.getParent()) {
            Path candidate = dir.resolve("seam/vectors");
            if (Files.isDirectory(candidate)) {
                return candidate;
            }
        }
        throw new IllegalStateException(
            "could not locate seam/vectors from " + Path.of("").toAbsolutePath());
    }

    @SuppressWarnings("unchecked")
    private static Map<String, Map<String, Object>> loadVectors() throws IOException {
        Path dir = vectorsDir();
        Map<String, Map<String, Object>> vectors = new TreeMap<>();
        try (DirectoryStream<Path> files = Files.newDirectoryStream(dir, "*.json")) {
            for (Path p : files) {
                Map<String, Object> v = (Map<String, Object>) CanonicalJson.decode(Files.readAllBytes(p));
                String stem = p.getFileName().toString().replaceFirst("\\.json$", "");
                assertEquals(stem, v.get("name"), "file/vector name mismatch for " + p);
                vectors.put(stem, v);
            }
        }
        return vectors;
    }

    @Test
    void census() throws IOException {
        Set<String> present = loadVectors().keySet();
        for (String name : CENSUS) {
            assertEquals(true, present.contains(name), "missing vector: " + name);
        }
        for (String name : present) {
            assertEquals(true, CENSUS.contains(name), "vector outside the closed census: " + name);
        }
    }

    @TestFactory
    Stream<DynamicTest> vectors() throws IOException {
        return loadVectors().entrySet().stream()
            .map(e -> DynamicTest.dynamicTest(e.getKey(), () -> checkVector(e.getKey(), e.getValue())));
    }

    @SuppressWarnings("unchecked")
    private static void checkVector(String name, Map<String, Object> vector) throws IOException {
        Object wantDecoded = vector.get("decoded");
        String pinnedHex = (String) vector.get("frame_hex");
        String expect = (String) vector.get("expect");
        byte[] frame = HexFormat.of().parseHex(pinnedHex);

        // 1+2. Decode the fixture's frame; the body must equal the fixture's `decoded` form.
        ReadableByteChannel in = Channels.newChannel(new java.io.ByteArrayInputStream(frame));
        byte[] body = FrameCodec.readFrame(in);
        assertNotNull(body, name + ": frame must decode a body");
        Object decoded = CanonicalJson.decode(body);
        assertEquals(wantDecoded, decoded, name + ": decoded value differs from the fixture's `decoded`");

        // 3. Canonical re-encode reproduces the pinned bytes, length prefix included.
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        FrameCodec.writeFrame(Channels.newChannel(out), CanonicalJson.encode(decoded));
        assertEquals(pinnedHex, HexFormat.of().formatHex(out.toByteArray()),
            name + ": re-encoded frame differs from the pinned bytes");

        // 4. Pinned validation behavior.
        String refusal = validate(decoded);
        switch (expect) {
            case "roundtrip", "decode_ok" ->
                assertNull(refusal, name + ": must be accepted but validation refused it: " + refusal);
            case "refuse" -> {
                String wanted = (String) vector.get("refusal");
                assertNotNull(refusal, name + ": must be refused (" + wanted + ") but validation accepted it");
                assertEquals(wanted, refusal, name + ": refusal mismatch");
            }
            default -> fail(name + ": unknown expect " + expect);
        }
    }

    /** V-5's required-field rules and nothing more (ported from RoundTrip.java, T005:
     * "mirroring the TASK-0007 harness obligations against THIS codec"). Enum membership is
     * deliberately not checked: V-2 requires an unrecognized value to decode and be
     * classified above the codec (err_unknown_origin). @return null when acceptable, else a
     * stable refusal string. */
    private static String validate(Object message) {
        if (!(message instanceof Map<?, ?> msg)) {
            return "malformed:message is not a JSON object";
        }
        for (String f : List.of("protocol", "message", "session", "seq", "body", "world_time", "payload")) {
            if (!msg.containsKey(f)) {
                return "missing_required_field:" + f;
            }
        }
        if (!(msg.get("message") instanceof String kind)) {
            return "malformed:message is not a string";
        }
        if (!(msg.get("payload") instanceof Map<?, ?> payload)) {
            return "malformed:payload is not a JSON object";
        }
        List<String> required = REQUIRED_PAYLOAD.get(kind);
        if (required == null) {
            return "malformed:unknown message type \"" + kind + "\"";
        }
        for (String f : required) {
            if (!payload.containsKey(f)) {
                return "missing_required_field:payload." + f;
            }
        }
        if (kind.equals("percept")) {
            if (!(payload.get("provenance") instanceof Map<?, ?> prov)) {
                return "malformed:payload.provenance is not a JSON object";
            }
            for (String f : List.of("origin", "source", "observed_at", "received_at")) {
                if (!prov.containsKey(f)) {
                    return "missing_required_field:payload.provenance." + f;
                }
            }
        }
        return null;
    }
}
