package dev.kithcraft.mod.percept;

import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Objects;

/**
 * §4.4 {@code sound}: heard, cause unknown. R-8 verified a native hearing hook — MC 26.2's
 * {@code net.minecraft.world.level.gameevent} package ({@code GameEvent}/{@code
 * GameEventListener}/{@code DynamicGameEventListener}/{@code EntityPositionSource}, the same
 * public API vanilla's Warden uses via {@code VibrationSystem.User}) lets a mod attach a
 * listener to an entity and learn a nearby event's direction/distance, with no Mixin
 * (specs/012-vendor-conformance/research/r8-hearing-hook.md). {@code sound} is therefore
 * declared in the manifest (T006).
 *
 * <p>This class only composes the wire content from an already-resolved {@code (sound_kind,
 * bearing, distance, descriptor)} — attaching a live {@code GameEventListener} to a villager
 * entity is Phase 3/4 wiring, once a live body exists to attach it to (ponytail: same shape as
 * {@link Sightings.HostileSource}; ceiling is unit-scoped Phase 2, add live listener wiring
 * when a body is actually connected to a world).
 *
 * <p>Normative (§4.4): the content MUST NOT name the thing that made the sound. This method's
 * signature enforces that structurally — there is no thing-typed parameter to leak.
 */
public final class Sounds {
    private Sounds() {}

    public static Map<String, Object> content(String soundKind, String bearing, String distance, String descriptor) {
        Objects.requireNonNull(soundKind, "soundKind");
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("sound_kind", soundKind);
        m.put("bearing", bearing);
        m.put("distance", distance);
        m.put("descriptor", descriptor);
        return m;
    }

    /** §4.4's provenance is always {@code heard}, direct — the body is its own witness to the
     * sound (never the sound's cause). {@code place} is typically {@code null} at the call
     * site: a sound gives a direction, not a location. */
    public static Provenance provenance(Long observedAt, long receivedAt) {
        return Provenance.direct(Provenance.Origin.HEARD, observedAt, receivedAt);
    }
}
