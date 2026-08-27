package dev.kithcraft.mod.cast;

import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import net.minecraft.resources.Identifier;
import net.minecraft.util.datafix.DataFixTypes;
import net.minecraft.world.level.saveddata.SavedData;
import net.minecraft.world.level.saveddata.SavedDataType;

/**
 * Persists {@link Cast} through the world save (T002), the exact pattern {@code
 * dev.kithcraft.mod.tokens.TokenRegistryData} established for {@code TokenRegistry}: a
 * codec-described {@link SavedDataType}, fetched from a {@code ServerLevel}'s save storage via
 * {@code computeIfAbsent(TYPE)}. This class is the thin adapter only — the seeded/not-seeded
 * rule lives in the pure {@link Cast}.
 */
public final class CastData extends SavedData {

    private static final Codec<Cast.Snapshot> SNAPSHOT_CODEC = RecordCodecBuilder.create(i -> i.group(
        Codec.BOOL.fieldOf("seeded").forGetter(Cast.Snapshot::seeded)
    ).apply(i, Cast.Snapshot::new));

    public static final Codec<CastData> CODEC = SNAPSHOT_CODEC.xmap(CastData::new, CastData::snapshot);

    public static final SavedDataType<CastData> TYPE = new SavedDataType<>(
        Identifier.fromNamespaceAndPath("kithcraft", "cast"),
        CastData::new, CODEC, DataFixTypes.LEVEL);

    private final Cast cast;

    public CastData() {
        this.cast = new Cast();
    }

    private CastData(Cast.Snapshot snapshot) {
        this.cast = Cast.restore(snapshot);
    }

    private Cast.Snapshot snapshot() {
        return cast.snapshot();
    }

    public boolean isSeeded() {
        return cast.isSeeded();
    }

    public void markSeeded() {
        cast.markSeeded();
        setDirty();
    }
}
