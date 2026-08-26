package dev.kithcraft.mod.percept;

import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Objects;

/** §2.4 place: an opaque token plus a vendor-composed, prose-only descriptor. No
 * coordinates ever cross the seam (AR-4). */
public record Place(String place, String descriptor) {
    public Place {
        Objects.requireNonNull(place, "place");
    }

    public Map<String, Object> toPayload() {
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("place", place);
        m.put("descriptor", descriptor);
        return m;
    }
}
