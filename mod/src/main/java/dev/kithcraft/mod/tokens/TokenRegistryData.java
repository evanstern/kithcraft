package dev.kithcraft.mod.tokens;

import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import net.minecraft.resources.Identifier;
import net.minecraft.util.datafix.DataFixTypes;
import net.minecraft.world.level.saveddata.SavedData;
import net.minecraft.world.level.saveddata.SavedDataType;

import java.util.Map;
import java.util.Optional;

/**
 * Persists {@link TokenRegistry} through the world save (T008).
 *
 * <p>MC 26.2 has no {@code PersistentState} class — checked directly against the
 * unobfuscated 26.2 jar ({@code minecraft-merged.jar}; no match for {@code PersistentState}
 * anywhere in it). The equivalent at this version is {@link SavedData} (an abstract,
 * dirty-flagged base) registered as a codec-described {@link SavedDataType} and fetched from
 * a {@code ServerLevel}'s {@code SavedDataStorage} via {@code computeIfAbsent(TYPE)}. This is
 * recorded in specs/009-fabric-mod-skeleton/research/versions.md.
 *
 * <p>This class is the thin adapter only: every issue/resolve/retire rule lives in the pure,
 * server-free {@link TokenRegistry}. This class exists solely to get a {@link
 * TokenRegistry.Snapshot} onto disk (as NBT, via the codec) and back, and to flag the vanilla
 * save pipeline dirty on every mutation so a normal world save persists it — no bespoke I/O.
 */
public final class TokenRegistryData extends SavedData {

    private static final Codec<TokenRegistry.Entry> ENTRY_CODEC = RecordCodecBuilder.create(i -> i.group(
        Codec.STRING.fieldOf("referent").forGetter(TokenRegistry.Entry::referent),
        Codec.BOOL.fieldOf("retired").forGetter(TokenRegistry.Entry::retired)
    ).apply(i, TokenRegistry.Entry::new));

    private static final Codec<TokenRegistry.Snapshot> SNAPSHOT_CODEC = RecordCodecBuilder.create(i -> i.group(
        Codec.LONG.fieldOf("next_sequence").forGetter(TokenRegistry.Snapshot::nextSequence),
        Codec.unboundedMap(Codec.STRING, ENTRY_CODEC).fieldOf("entries").forGetter(TokenRegistry.Snapshot::entries)
    ).apply(i, TokenRegistry.Snapshot::new));

    /** {@code xmap}s the snapshot codec onto this adapter directly, so the codec only ever
     * has to know the pure {@link TokenRegistry.Snapshot} shape. */
    public static final Codec<TokenRegistryData> CODEC =
        SNAPSHOT_CODEC.xmap(TokenRegistryData::new, TokenRegistryData::snapshot);

    /** Not one of vanilla's own {@code SAVED_DATA_*} fix types (those are reserved for
     * vanilla's systems) — {@link DataFixTypes#LEVEL} is the generic per-level choice for
     * mod-owned data, the same one third-party persistent state has conventionally used. */
    public static final SavedDataType<TokenRegistryData> TYPE = new SavedDataType<>(
        Identifier.fromNamespaceAndPath("kithcraft", "token_registry"),
        TokenRegistryData::new, CODEC, DataFixTypes.LEVEL);

    private final TokenRegistry registry;

    public TokenRegistryData() {
        this.registry = new TokenRegistry();
    }

    private TokenRegistryData(TokenRegistry.Snapshot snapshot) {
        this.registry = TokenRegistry.restore(snapshot);
    }

    private TokenRegistry.Snapshot snapshot() {
        return registry.snapshot();
    }

    public String issue(TokenRegistry.TokenType type, String referent) {
        String token = registry.issue(type, referent);
        setDirty();
        return token;
    }

    public Optional<String> resolve(String token) {
        return registry.resolve(token);
    }

    public void retire(String token) {
        registry.retire(token);
        setDirty();
    }

    /** All tokens that currently resolve — used by the dev-server restart observation
     * (T009) and general diagnostics. */
    public Map<String, String> liveEntries() {
        return registry.liveEntries();
    }
}
