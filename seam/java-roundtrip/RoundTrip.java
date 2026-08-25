// RoundTrip — the JVM half of the seam wire's executable pinning.
//
// Run:  java RoundTrip.java ../vectors
//
// This is a throwaway-grade proof of docs/design/seam-wire-v0.md, not the transport:
// V1 (TASK-0009) builds that. Its value is entirely in being a SECOND, INDEPENDENT
// implementation — it shares no code with the Go harness on purpose, because two
// implementations agreeing is the only evidence that the spec, rather than one
// codebase's habits, is what the vectors pin.
//
// JDK stdlib only. The JSON handling below is hand-rolled because the canonical form
// (§2.4) is a small closed subset — objects, arrays, strings, integers, the three
// literals — and pulling a JSON library in would mean a build system, which this task
// must not introduce.

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HexFormat;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Set;
import java.util.TreeMap;

public final class RoundTrip {

    /** seam-wire-v0.md §2.5. */
    private static final int MAX_FRAME_BODY = 1 << 20;

    /** contracts/vectors.md's closed list. Extras are scope creep; gaps are holes. */
    private static final List<String> CENSUS = List.of(
        "percept_sighting", "percept_observation", "percept_sound", "percept_speech",
        "percept_told_fact", "percept_text", "percept_act_result", "percept_self_state",
        "percept_change_report",
        "intent", "intent_ack", "cancel",
        "session_open", "session_close",
        "err_missing_provenance", "err_unknown_origin", "intent_ack_refused");

    private static int passed = 0;
    private static final List<String> failures = new ArrayList<>();

    public static void main(String[] args) throws IOException {
        Path dir = Path.of(args.length > 0 ? args[0] : "../vectors");
        if (!Files.isDirectory(dir)) {
            System.err.println("no vector directory at " + dir.toAbsolutePath());
            System.exit(2);
        }

        Map<String, Map<String, Object>> vectors = new TreeMap<>();
        try (DirectoryStream<Path> files = Files.newDirectoryStream(dir, "*.json")) {
            for (Path p : files) {
                Object parsed = Json.parse(Files.readString(p, StandardCharsets.UTF_8));
                @SuppressWarnings("unchecked")
                Map<String, Object> v = (Map<String, Object>) parsed;
                String stem = p.getFileName().toString().replaceFirst("\\.json$", "");
                if (!stem.equals(v.get("name"))) {
                    fail(stem, "declares name " + v.get("name") + "; file and vector name must agree");
                    continue;
                }
                vectors.put(stem, v);
            }
        }

        checkCensus(vectors.keySet());
        for (Map.Entry<String, Map<String, Object>> e : vectors.entrySet()) {
            try {
                checkVector(e.getKey(), e.getValue());
            } catch (RuntimeException ex) {
                fail(e.getKey(), ex.toString());
            }
        }
        checkFramingRefusals();
        checkReceiverAcceptsNonCanonical();

        System.out.println();
        System.out.printf("java-roundtrip: %d passed, %d failed, over %d vectors%n",
            passed, failures.size(), vectors.size());
        if (!failures.isEmpty()) {
            System.out.println();
            failures.forEach(f -> System.out.println("  FAIL " + f));
            System.exit(1);
        }
    }

    // ---------------------------------------------------------------- checks

    private static void checkCensus(Set<String> present) {
        for (String name : CENSUS) {
            if (present.contains(name)) {
                pass("census/" + name);
            } else {
                fail("census/" + name, "missing — contracts/vectors.md requires it");
            }
        }
        for (String name : present) {
            if (!CENSUS.contains(name)) {
                fail("census/" + name, "outside the closed census — extras are scope creep");
            }
        }
    }

    private static void checkVector(String name, Map<String, Object> vector) {
        Object want = vector.get("decoded");
        String pinnedHex = (String) vector.get("frame_hex");
        String expect = (String) vector.get("expect");

        if (pinnedHex == null || !pinnedHex.equals(pinnedHex.toLowerCase())) {
            fail(name, "frame_hex must be present and lowercase (§6 fixture convention)");
            return;
        }
        byte[] frame = HexFormat.of().parseHex(pinnedHex);

        // 1. The frame decodes, and consumes exactly its own bytes.
        Frame decoded = decodeFrame(frame, 0);
        if (decoded.end != frame.length) {
            fail(name, "frame declares " + decoded.end + " bytes but the fixture carries " + frame.length);
            return;
        }
        pass(name + "/decode");

        // 2. Meaning: structural equality against the pinned decoded form.
        //    Objects compare as unordered maps, arrays in order, numbers by value —
        //    which Map/List/Long equals() already give, provided nothing above ever
        //    relied on iteration order.
        if (!structurallyEqual(decoded.value, want)) {
            fail(name + "/meaning", "decoded value differs from the declared form\n"
                + "        got: " + Json.canonical(decoded.value) + "\n"
                + "       want: " + Json.canonical(want));
        } else {
            pass(name + "/meaning");
        }

        // 3. Bytes: canonical re-encode equals the pinned frame, length word included.
        byte[] reencoded = encodeFrame(decoded.value);
        String gotHex = HexFormat.of().formatHex(reencoded);
        if (!gotHex.equals(pinnedHex)) {
            fail(name + "/bytes", "re-encoded frame differs from the pinned bytes\n"
                + "       " + hexDiff(pinnedHex, gotHex));
        } else {
            pass(name + "/bytes");
        }

        // 4. Validation behaves as the vector pins.
        String refusal = validate(decoded.value);
        switch (expect) {
            case "roundtrip", "decode_ok" -> {
                if (refusal != null) {
                    fail(name + "/accept", "must be accepted but validation refused it: " + refusal);
                } else {
                    pass(name + "/accept");
                }
            }
            case "refuse" -> {
                String wanted = (String) vector.get("refusal");
                if (refusal == null) {
                    fail(name + "/refuse", "must be refused (" + wanted + ") but validation accepted it");
                } else if (!refusal.equals(wanted)) {
                    fail(name + "/refuse", "refusal mismatch: got " + refusal + ", want " + wanted);
                } else {
                    pass(name + "/refuse");
                }
            }
            default -> fail(name, "unknown expect " + expect);
        }
    }

    /** Frame-layer rules no fixture can carry, because a malformed frame is not a frame. */
    private static void checkFramingRefusals() {
        refuses("framing/zero-length", new byte[] {0, 0, 0, 0});
        refuses("framing/short-header", new byte[] {0, 0, 1});
        refuses("framing/truncated-body", new byte[] {0, 0, 0, 8, '{', '}'});
        refuses("framing/over-cap", new byte[] {0, 0x20, 0, 1});

        // Back-to-back frames: the next length word begins at the byte after the
        // previous body's last byte (§2.1), with no delimiter between them.
        byte[] one = encodeFrame(Map.of("a", 1L));
        byte[] two = encodeFrame(Map.of("b", 2L));
        byte[] stream = new byte[one.length + two.length];
        System.arraycopy(one, 0, stream, 0, one.length);
        System.arraycopy(two, 0, stream, one.length, two.length);
        Frame f1 = decodeFrame(stream, 0);
        Frame f2 = decodeFrame(stream, f1.end);
        if (f1.end == one.length && f2.end == stream.length
                && structurallyEqual(f1.value, Map.of("a", 1L))
                && structurallyEqual(f2.value, Map.of("b", 2L))) {
            pass("framing/back-to-back");
        } else {
            fail("framing/back-to-back", "frames did not decode independently from one stream");
        }
    }

    /** §2.4's asymmetry: senders emit canonical, receivers accept any conforming JSON. */
    private static void checkReceiverAcceptsNonCanonical() {
        Object v = Json.parse("{ \"b\" : 2,\n  \"a\" : 1 }");
        String canon = Json.canonical(v);
        if (canon.equals("{\"a\":1,\"b\":2}")) {
            pass("receiver/accepts-non-canonical");
        } else {
            fail("receiver/accepts-non-canonical", "canonical form of the same value = " + canon);
        }
    }

    private static void refuses(String label, byte[] frame) {
        try {
            decodeFrame(frame, 0);
            fail(label, "expected refusal, got acceptance");
        } catch (RuntimeException expected) {
            pass(label);
        }
    }

    // ------------------------------------------------------------- the wire

    private record Frame(Object value, int end) {}

    /** §2.1/§2.3: 4-byte big-endian length, then exactly that many bytes of UTF-8 JSON. */
    private static Frame decodeFrame(byte[] buf, int offset) {
        if (buf.length - offset < 4) {
            throw new IllegalArgumentException("truncated frame: fewer than 4 bytes for the length word");
        }
        long n = ((long) (buf[offset] & 0xFF) << 24)
               | ((long) (buf[offset + 1] & 0xFF) << 16)
               | ((long) (buf[offset + 2] & 0xFF) << 8)
               | (buf[offset + 3] & 0xFF);
        if (n == 0) {
            throw new IllegalArgumentException("malformed frame: length 0 (an empty body is not a JSON value)");
        }
        if (n > MAX_FRAME_BODY) {
            throw new IllegalArgumentException("frame body " + n + " exceeds the 1 MiB cap");
        }
        if (buf.length - offset - 4 < n) {
            throw new IllegalArgumentException("truncated frame: body declared " + n + " bytes");
        }
        int start = offset + 4, end = (int) (start + n);
        byte[] body = java.util.Arrays.copyOfRange(buf, start, end);
        if (body.length >= 3 && (body[0] & 0xFF) == 0xEF && (body[1] & 0xFF) == 0xBB && (body[2] & 0xFF) == 0xBF) {
            throw new IllegalArgumentException("body carries a BOM (C-1)");
        }
        // Decoding strictly, so invalid UTF-8 raises rather than becoming U+FFFD.
        String text;
        try {
            text = StandardCharsets.UTF_8.newDecoder()
                .onMalformedInput(java.nio.charset.CodingErrorAction.REPORT)
                .onUnmappableCharacter(java.nio.charset.CodingErrorAction.REPORT)
                .decode(java.nio.ByteBuffer.wrap(body))
                .toString();
        } catch (java.nio.charset.CharacterCodingException e) {
            throw new IllegalArgumentException("body is not valid UTF-8: " + e.getMessage());
        }
        return new Frame(Json.parse(text), end);
    }

    private static byte[] encodeFrame(Object value) {
        byte[] body = Json.canonical(value).getBytes(StandardCharsets.UTF_8);
        if (body.length > MAX_FRAME_BODY) {
            throw new IllegalArgumentException("frame body " + body.length + " exceeds the 1 MiB cap");
        }
        byte[] frame = new byte[4 + body.length];
        frame[0] = (byte) (body.length >>> 24);
        frame[1] = (byte) (body.length >>> 16);
        frame[2] = (byte) (body.length >>> 8);
        frame[3] = (byte) body.length;
        System.arraycopy(body, 0, frame, 4, body.length);
        return frame;
    }

    // -------------------------------------------------------- message rules

    /** Required payload members per message type (✔ in the protocol's tables). */
    private static final Map<String, List<String>> REQUIRED_PAYLOAD = Map.of(
        "percept", List.of("percept_id", "percept_type", "urgency", "provenance", "place", "content"),
        "intent", List.of("intent_id", "verb", "target", "reason"),
        "intent_ack", List.of("intent_id", "accepted", "reason_code"),
        "cancel", List.of("intent_id"),
        "session_open", List.of("time_unit", "capabilities"),
        "session_close", List.of("reason", "detail"));

    /**
     * V-5's required-field rules, and nothing more. Enum membership is deliberately
     * NOT checked: V-2 requires an unrecognized value to decode and be classified
     * above the codec, which is exactly what err_unknown_origin pins.
     *
     * @return null when acceptable, else a stable refusal string.
     */
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
        // EH-2a: a percept's provenance is required, and a provenance missing its own
        // required members is as unusable as none at all. Neither may be defaulted.
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

    // ------------------------------------------------------------ utilities

    /** Objects unordered, arrays in order, numbers by value (§6's second assertion). */
    private static boolean structurallyEqual(Object a, Object b) {
        if (a instanceof Map<?, ?> ma && b instanceof Map<?, ?> mb) {
            if (!ma.keySet().equals(mb.keySet())) {
                return false;
            }
            for (Object k : ma.keySet()) {
                if (!structurallyEqual(ma.get(k), mb.get(k))) {
                    return false;
                }
            }
            return true;
        }
        if (a instanceof List<?> la && b instanceof List<?> lb) {
            if (la.size() != lb.size()) {
                return false;
            }
            for (int i = 0; i < la.size(); i++) {
                if (!structurallyEqual(la.get(i), lb.get(i))) {
                    return false;
                }
            }
            return true;
        }
        return Objects.equals(a, b);
    }

    /** Hex is byte-aligned, so a mismatch offset is computable by hand (§6). */
    private static String hexDiff(String want, String got) {
        int n = Math.min(want.length(), got.length());
        for (int i = 0; i + 1 < n; i += 2) {
            if (!want.startsWith(got.substring(i, i + 2), i)) {
                return "first difference at byte " + (i / 2)
                    + ": want " + want.substring(i, i + 2) + ", got " + got.substring(i, i + 2)
                    + "\n        want: …" + window(want, i) + "…"
                    + "\n         got: …" + window(got, i) + "…";
            }
        }
        return "frames agree for " + (n / 2) + " bytes then differ in length: want "
            + (want.length() / 2) + " bytes, got " + (got.length() / 2);
    }

    private static String window(String s, int i) {
        return s.substring(Math.max(0, i - 16), Math.min(s.length(), i + 16));
    }

    private static void pass(String label) {
        passed++;
    }

    private static void fail(String label, String why) {
        failures.add(label + ": " + why);
    }

    // ===================================================================== JSON

    /**
     * A minimal RFC 8259 reader and a canonical-form writer (§2.4, C-1..C-10).
     * Numbers become {@code Long} because C-6 makes every v0 number an integer and
     * requires a receiver to refuse one outside signed-64-bit range rather than lose
     * it to a double.
     */
    static final class Json {

        private final String s;
        private int i;

        private Json(String s) {
            this.s = s;
        }

        static Object parse(String text) {
            Json p = new Json(text);
            p.skipWhitespace();
            Object v = p.value();
            p.skipWhitespace();
            if (p.i != text.length()) {
                throw new IllegalArgumentException("trailing content after the message at offset " + p.i);
            }
            return v;
        }

        private Object value() {
            if (i >= s.length()) {
                throw new IllegalArgumentException("unexpected end of input");
            }
            return switch (s.charAt(i)) {
                case '{' -> object();
                case '[' -> array();
                case '"' -> string();
                case 't' -> literal("true", Boolean.TRUE);
                case 'f' -> literal("false", Boolean.FALSE);
                case 'n' -> literal("null", null);
                default -> number();
            };
        }

        private Map<String, Object> object() {
            expect('{');
            // LinkedHashMap keeps source order for readable diagnostics; comparison
            // and canonical output never depend on it.
            Map<String, Object> out = new LinkedHashMap<>();
            skipWhitespace();
            if (peek() == '}') {
                i++;
                return out;
            }
            while (true) {
                skipWhitespace();
                String key = string();
                if (out.containsKey(key)) {
                    // C-4: last-wins would silently pick a value the sender did not mean.
                    throw new IllegalArgumentException("duplicate key \"" + key + "\" (C-4)");
                }
                skipWhitespace();
                expect(':');
                skipWhitespace();
                out.put(key, value());
                skipWhitespace();
                char c = next();
                if (c == '}') {
                    return out;
                }
                if (c != ',') {
                    throw new IllegalArgumentException("expected ',' or '}' at offset " + (i - 1));
                }
            }
        }

        private List<Object> array() {
            expect('[');
            List<Object> out = new ArrayList<>();
            skipWhitespace();
            if (peek() == ']') {
                i++;
                return out;
            }
            while (true) {
                skipWhitespace();
                out.add(value());
                skipWhitespace();
                char c = next();
                if (c == ']') {
                    return out;
                }
                if (c != ',') {
                    throw new IllegalArgumentException("expected ',' or ']' at offset " + (i - 1));
                }
            }
        }

        private String string() {
            expect('"');
            StringBuilder sb = new StringBuilder();
            while (true) {
                char c = next();
                if (c == '"') {
                    return sb.toString();
                }
                if (c < 0x20) {
                    throw new IllegalArgumentException("unescaped control character in string at offset " + (i - 1));
                }
                if (c != '\\') {
                    sb.append(c);
                    continue;
                }
                char esc = next();
                switch (esc) {
                    case '"' -> sb.append('"');
                    case '\\' -> sb.append('\\');
                    case '/' -> sb.append('/');
                    case 'b' -> sb.append('\b');
                    case 'f' -> sb.append('\f');
                    case 'n' -> sb.append('\n');
                    case 'r' -> sb.append('\r');
                    case 't' -> sb.append('\t');
                    case 'u' -> {
                        if (i + 4 > s.length()) {
                            throw new IllegalArgumentException("truncated \\u escape");
                        }
                        sb.append((char) Integer.parseInt(s.substring(i, i + 4), 16));
                        i += 4;
                    }
                    default -> throw new IllegalArgumentException("invalid escape \\" + esc);
                }
            }
        }

        private Long number() {
            int start = i;
            if (peek() == '-') {
                i++;
            }
            while (i < s.length() && s.charAt(i) >= '0' && s.charAt(i) <= '9') {
                i++;
            }
            String lexeme = s.substring(start, i);
            if (i < s.length() && (s.charAt(i) == '.' || s.charAt(i) == 'e' || s.charAt(i) == 'E')) {
                throw new IllegalArgumentException("number " + lexeme + "… is not an integer (C-6)");
            }
            if (lexeme.isEmpty() || lexeme.equals("-")) {
                throw new IllegalArgumentException("expected a value at offset " + start);
            }
            try {
                return Long.parseLong(lexeme);
            } catch (NumberFormatException e) {
                throw new IllegalArgumentException("number " + lexeme + " is not a signed 64-bit integer (C-6)");
            }
        }

        private Object literal(String word, Object v) {
            if (!s.startsWith(word, i)) {
                throw new IllegalArgumentException("invalid literal at offset " + i);
            }
            i += word.length();
            return v;
        }

        private void skipWhitespace() {
            while (i < s.length()) {
                char c = s.charAt(i);
                if (c == ' ' || c == '\t' || c == '\n' || c == '\r') {
                    i++;
                } else {
                    return;
                }
            }
        }

        private char peek() {
            if (i >= s.length()) {
                throw new IllegalArgumentException("unexpected end of input");
            }
            return s.charAt(i);
        }

        private char next() {
            char c = peek();
            i++;
            return c;
        }

        private void expect(char c) {
            if (next() != c) {
                throw new IllegalArgumentException("expected '" + c + "' at offset " + (i - 1));
            }
        }

        // ------------------------------------------------------ canonical out

        static String canonical(Object v) {
            StringBuilder sb = new StringBuilder();
            write(sb, v);
            return sb.toString();
        }

        private static void write(StringBuilder sb, Object v) {
            switch (v) {
                case null -> sb.append("null");                                 // C-10
                case Boolean b -> sb.append(b ? "true" : "false");              // C-10
                case Long n -> sb.append(n.longValue());                        // C-6
                case Integer n -> sb.append(n.longValue());
                case String str -> writeString(sb, str);
                case List<?> list -> {
                    sb.append('[');
                    for (int k = 0; k < list.size(); k++) {                     // C-5
                        if (k > 0) {
                            sb.append(',');
                        }
                        write(sb, list.get(k));
                    }
                    sb.append(']');
                }
                case Map<?, ?> map -> {
                    List<String> keys = new ArrayList<>();
                    map.keySet().forEach(k -> keys.add((String) k));
                    // C-3: ascending by the key's UTF-8 bytes. Every v0 key is ASCII,
                    // so String's natural (UTF-16 code-unit) order coincides; sorting
                    // the encoded bytes keeps it true if a later version adds a key
                    // above U+FFFF, where the two orders diverge.
                    keys.sort((x, y) -> java.util.Arrays.compareUnsigned(
                        x.getBytes(StandardCharsets.UTF_8), y.getBytes(StandardCharsets.UTF_8)));
                    sb.append('{');
                    for (int k = 0; k < keys.size(); k++) {
                        if (k > 0) {
                            sb.append(',');                                     // C-2
                        }
                        writeString(sb, keys.get(k));
                        sb.append(':');
                        write(sb, map.get(keys.get(k)));
                    }
                    sb.append('}');
                }
                default -> throw new IllegalArgumentException(
                    "value of type " + v.getClass().getName() + " is not representable on this wire");
            }
        }

        /** C-7 and C-8: RFC-required escapes only, literal UTF-8, no lone surrogate. */
        private static void writeString(StringBuilder sb, String str) {
            sb.append('"');
            for (int k = 0; k < str.length(); k++) {
                char c = str.charAt(k);
                switch (c) {
                    case '"' -> sb.append("\\\"");
                    case '\\' -> sb.append("\\\\");
                    case '\b' -> sb.append("\\b");
                    case '\t' -> sb.append("\\t");
                    case '\n' -> sb.append("\\n");
                    case '\f' -> sb.append("\\f");
                    case '\r' -> sb.append("\\r");
                    default -> {
                        if (c < 0x20) {
                            sb.append(String.format("\\u%04x", (int) c));       // lowercase hex
                        } else if (Character.isHighSurrogate(c)) {
                            // C-8: a UTF-16-native runtime can hold an unpaired
                            // surrogate that no valid UTF-8 can represent. Transcoding
                            // it to U+FFFD would silently alter an utterance, so refuse.
                            if (k + 1 >= str.length() || !Character.isLowSurrogate(str.charAt(k + 1))) {
                                throw new IllegalArgumentException("lone high surrogate in string (C-8)");
                            }
                            sb.append(c).append(str.charAt(++k));
                        } else if (Character.isLowSurrogate(c)) {
                            throw new IllegalArgumentException("lone low surrogate in string (C-8)");
                        } else {
                            sb.append(c);   // literal UTF-8 once encoded; / is not escaped
                        }
                    }
                }
            }
            sb.append('"');
        }
    }

    private RoundTrip() {}
}
