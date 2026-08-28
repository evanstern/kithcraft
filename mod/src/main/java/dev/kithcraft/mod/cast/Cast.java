package dev.kithcraft.mod.cast;

import java.util.List;
import java.util.Objects;

/**
 * The three-member cast (FR-002, card AC #6): profession × biome variant plus a name, one
 * per member. Pure core — no Minecraft/Fabric types — so it is unit-testable without a live
 * server; {@link CastData} is the thin adapter that persists whether seeding has already
 * happened, mirroring {@code dev.kithcraft.mod.tokens.TokenRegistry}/{@code
 * TokenRegistryData}'s split.
 *
 * <p>Identity survival across a restart does not need to track which villagers were spawned —
 * the spawned {@code Villager} entities are ordinary world entities and persist through the
 * normal save/load path on their own (name, profession, and biome variant are vanilla NBT
 * fields). All this class's persisted state has to prevent is spawning the cast a second time
 * on the next boot.
 *
 * <p>ponytail: a villager that dies leaves the {@code seeded} flag permanently true with no
 * replacement spawned — death/grief handling is V5's scope, not V3's; this is the accepted
 * ceiling until then.
 */
public final class Cast {

    /** One cast member's identity: display name, vanilla profession id, vanilla biome-variant
     * (VillagerType) id — both plain {@code minecraft:}-namespaced path strings, resolved
     * against the live registries only where a Minecraft type is actually needed ({@link
     * CastSeeder}). */
    public record Member(String name, String profession, String biomeVariant) {
        public Member {
            Objects.requireNonNull(name, "name");
            Objects.requireNonNull(profession, "profession");
            Objects.requireNonNull(biomeVariant, "biomeVariant");
        }
    }

    /** The cast, fixed: three distinct profession × biome variant pairings (card AC #6). */
    public static final List<Member> MEMBERS = List.of(
        new Member("Aldric", "armorer", "plains"),
        new Member("Petra", "farmer", "desert"),
        new Member("Yenna", "fisherman", "taiga")
    );

    /** Persisted state: has the cast already been spawned once. */
    public record Snapshot(boolean seeded) {
    }

    private boolean seeded;

    public Cast() {
        this.seeded = false;
    }

    private Cast(boolean seeded) {
        this.seeded = seeded;
    }

    public boolean isSeeded() {
        return seeded;
    }

    /** Idempotent: calling this more than once leaves the state exactly as if it were called
     * once (seeding idempotence, T002's unit test). */
    public void markSeeded() {
        this.seeded = true;
    }

    public Snapshot snapshot() {
        return new Snapshot(seeded);
    }

    public static Cast restore(Snapshot snapshot) {
        Objects.requireNonNull(snapshot, "snapshot");
        return new Cast(snapshot.seeded());
    }
}
