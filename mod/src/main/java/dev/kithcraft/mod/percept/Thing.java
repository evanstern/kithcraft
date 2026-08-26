package dev.kithcraft.mod.percept;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;

/** §2.5 thing: an opaque {@code kind} token (AR-2) plus {@code roles}/{@code descriptor}
 * (AR-3) carrying the meaning — never a native type name, never coordinates. */
public record Thing(String thingId, String kind, List<String> roles, String descriptor,
                     String body, int count) {
    public Thing {
        Objects.requireNonNull(kind, "kind");
        roles = roles == null ? List.of() : List.copyOf(roles);
    }

    /** A single, un-individuated thing (the common case): no {@code thing_id}, no body,
     * {@code count} 1. */
    public static Thing of(String thingId, String kind, List<String> roles, String descriptor) {
        return new Thing(thingId, kind, roles, descriptor, null, 1);
    }

    public Map<String, Object> toPayload() {
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("thing_id", thingId);
        m.put("kind", kind);
        m.put("roles", roles);
        m.put("descriptor", descriptor);
        m.put("body", body);
        m.put("count", (long) count);
        return m;
    }
}
