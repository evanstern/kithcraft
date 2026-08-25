package dev.kithcraft.mod.wire;

import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonToken;

import java.io.IOException;
import java.io.StringReader;
import java.nio.charset.CharacterCodingException;
import java.nio.charset.CodingErrorAction;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Canonical JSON for the seam wire (docs/design/seam-wire-v0.md §2.4, C-1..C-10).
 *
 * <p><b>Decode</b> uses Gson's {@link JsonReader} to tokenize, with the wire's stricter
 * rules layered on top of Gson's default (lenient-about-value) behaviour: no duplicate
 * keys (C-4), numbers must be plain integers in signed-64-bit range (C-6), no lone
 * surrogate in any string (C-8), no BOM and strict UTF-8 (C-1). Per §2.4's asymmetry,
 * decode is otherwise lenient about form — whitespace, key order — since receivers must
 * accept any conforming encoding of the same value.
 *
 * <p><b>Encode</b> is hand-written, not Gson's {@code JsonWriter}: Gson has no sorted-key
 * mode (C-3) and no integer-only numeric mode (C-6), so a thin wrapper cannot reach
 * canonical form. Recorded in specs/009-fabric-mod-skeleton/research/versions.md, verified
 * byte-exact against all 17 seam/vectors fixtures.
 *
 * <p>Values are the plain JDK shapes {@code Map<String,Object>} / {@code List<Object>} /
 * {@code Long} / {@code String} / {@code Boolean} / {@code null} — no library-specific
 * element types leak past this class.
 */
public final class CanonicalJson {
    private CanonicalJson() {}

    // ---------------------------------------------------------------- decode

    /** Decodes one frame body. Throws {@link IllegalArgumentException} for any violation
     * of C-1, C-4, C-6, or C-8, or any JSON syntax error. */
    public static Object decode(byte[] utf8) {
        if (utf8.length >= 3 && (utf8[0] & 0xFF) == 0xEF && (utf8[1] & 0xFF) == 0xBB && (utf8[2] & 0xFF) == 0xBF) {
            throw new IllegalArgumentException("body carries a BOM (C-1)");
        }
        String text;
        try {
            text = StandardCharsets.UTF_8.newDecoder()
                .onMalformedInput(CodingErrorAction.REPORT)
                .onUnmappableCharacter(CodingErrorAction.REPORT)
                .decode(java.nio.ByteBuffer.wrap(utf8))
                .toString();
        } catch (CharacterCodingException e) {
            throw new IllegalArgumentException("body is not valid UTF-8 (C-1): " + e.getMessage());
        }

        JsonReader reader = new JsonReader(new StringReader(text));
        Object value;
        try {
            value = readValue(reader);
            if (reader.peek() != JsonToken.END_DOCUMENT) {
                throw new IllegalArgumentException("trailing content after the message");
            }
        } catch (IOException e) {
            // Gson reports malformed JSON syntax via (Malformed)JsonException, itself an
            // IOException's checked sibling in older API surfaces; normalize to one
            // exception type for callers.
            throw new IllegalArgumentException("malformed JSON: " + e.getMessage(), e);
        }
        checkNoLoneSurrogates(value);
        return value;
    }

    private static Object readValue(JsonReader r) throws IOException {
        return switch (r.peek()) {
            case BEGIN_OBJECT -> readObject(r);
            case BEGIN_ARRAY -> readArray(r);
            case STRING -> r.nextString();
            case NUMBER -> readInteger(r);
            case BOOLEAN -> r.nextBoolean();
            case NULL -> {
                r.nextNull();
                yield null;
            }
            default -> throw new IllegalArgumentException("unexpected token " + r.peek());
        };
    }

    private static Map<String, Object> readObject(JsonReader r) throws IOException {
        Map<String, Object> out = new LinkedHashMap<>();
        r.beginObject();
        while (r.hasNext()) {
            String key = r.nextName();
            if (out.containsKey(key)) {
                throw new IllegalArgumentException("duplicate key \"" + key + "\" (C-4)");
            }
            out.put(key, readValue(r));
        }
        r.endObject();
        return out;
    }

    private static List<Object> readArray(JsonReader r) throws IOException {
        List<Object> out = new ArrayList<>();
        r.beginArray();
        while (r.hasNext()) {
            out.add(readValue(r));
        }
        r.endArray();
        return out;
    }

    /** C-6: every v0 number is an integer, representable as a signed 64-bit value. */
    private static Long readInteger(JsonReader r) throws IOException {
        String lexeme = r.nextString(); // Gson permits reading a NUMBER token as its raw text.
        if (lexeme.indexOf('.') >= 0 || lexeme.indexOf('e') >= 0 || lexeme.indexOf('E') >= 0) {
            throw new IllegalArgumentException("number " + lexeme + " is not an integer (C-6)");
        }
        try {
            return Long.parseLong(lexeme);
        } catch (NumberFormatException e) {
            throw new IllegalArgumentException("number " + lexeme + " is not a signed 64-bit integer (C-6)");
        }
    }

    /** C-8: a lone surrogate MUST be refused, never transcoded. */
    @SuppressWarnings("unchecked")
    private static void checkNoLoneSurrogates(Object v) {
        if (v instanceof String s) {
            for (int i = 0; i < s.length(); i++) {
                char c = s.charAt(i);
                if (Character.isHighSurrogate(c)) {
                    if (i + 1 >= s.length() || !Character.isLowSurrogate(s.charAt(i + 1))) {
                        throw new IllegalArgumentException("lone high surrogate in string (C-8)");
                    }
                    i++;
                } else if (Character.isLowSurrogate(c)) {
                    throw new IllegalArgumentException("lone low surrogate in string (C-8)");
                }
            }
        } else if (v instanceof Map<?, ?> m) {
            for (Object val : m.values()) {
                checkNoLoneSurrogates(val);
            }
        } else if (v instanceof List<?> l) {
            for (Object val : l) {
                checkNoLoneSurrogates(val);
            }
        }
    }

    // ---------------------------------------------------------------- encode

    /** Encodes canonical form (C-1..C-10). Input must be built from Map/List/Long/String/
     * Boolean/null; anything else throws. */
    public static byte[] encode(Object value) {
        StringBuilder sb = new StringBuilder();
        write(sb, value);
        return sb.toString().getBytes(StandardCharsets.UTF_8);
    }

    private static void write(StringBuilder sb, Object v) {
        switch (v) {
            case null -> sb.append("null"); // C-10
            case Boolean b -> sb.append(b ? "true" : "false"); // C-10
            case Long n -> sb.append(n.longValue()); // C-6
            case Integer n -> sb.append(n.longValue());
            case String str -> writeString(sb, str);
            case List<?> list -> {
                sb.append('[');
                for (int k = 0; k < list.size(); k++) { // C-5: array order untouched
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
                // C-3: ascending by the key's UTF-8 bytes.
                keys.sort((x, y) -> java.util.Arrays.compareUnsigned(
                    x.getBytes(StandardCharsets.UTF_8), y.getBytes(StandardCharsets.UTF_8)));
                sb.append('{');
                for (int k = 0; k < keys.size(); k++) {
                    if (k > 0) {
                        sb.append(','); // C-2: no insignificant whitespace
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

    /** C-7/C-8: RFC-required escapes only, literal UTF-8, no lone surrogate emitted. */
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
                        sb.append(String.format("\\u%04x", (int) c)); // lowercase hex
                    } else if (Character.isHighSurrogate(c)) {
                        if (k + 1 >= str.length() || !Character.isLowSurrogate(str.charAt(k + 1))) {
                            throw new IllegalArgumentException("lone high surrogate in string (C-8)");
                        }
                        sb.append(c).append(str.charAt(++k));
                    } else if (Character.isLowSurrogate(c)) {
                        throw new IllegalArgumentException("lone low surrogate in string (C-8)");
                    } else {
                        sb.append(c); // literal UTF-8 once encoded; '/' is not escaped
                    }
                }
            }
        }
        sb.append('"');
    }
}
