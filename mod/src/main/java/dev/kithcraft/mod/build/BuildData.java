package dev.kithcraft.mod.build;

import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import net.minecraft.resources.Identifier;
import net.minecraft.util.datafix.DataFixTypes;
import net.minecraft.world.level.saveddata.SavedData;
import net.minecraft.world.level.saveddata.SavedDataType;

import java.util.Optional;

/**
 * Persists {@link BuildCursor} through the world save (T009), the same adapter pattern {@code
 * BoardData}/{@code CastData}/{@code TokenRegistryData} established. {@code claimant_body=""}
 * is the "no build in progress yet" sentinel — the same empty-string-means-absent convention
 * {@link dev.kithcraft.mod.board.Board}'s own {@code claimedBy} snapshot already uses.
 */
public final class BuildData extends SavedData {

    private static final Codec<BuildCursor> CURSOR_CODEC = RecordCodecBuilder.create(i -> i.group(
        Codec.STRING.fieldOf("claimant_body").forGetter(BuildCursor::claimantBody),
        Codec.STRING.fieldOf("blueprint_text").forGetter(BuildCursor::blueprintText),
        Codec.INT.fieldOf("placed").forGetter(BuildCursor::placed)
    ).apply(i, BuildCursor::new));

    public static final Codec<BuildData> CODEC = CURSOR_CODEC.xmap(BuildData::new, BuildData::snapshot);

    public static final SavedDataType<BuildData> TYPE = new SavedDataType<>(
        Identifier.fromNamespaceAndPath("kithcraft", "build"),
        BuildData::new, CODEC, DataFixTypes.LEVEL);

    private BuildCursor cursor;

    public BuildData() {
        this.cursor = new BuildCursor("", "", 0);
    }

    private BuildData(BuildCursor cursor) {
        this.cursor = cursor;
    }

    private BuildCursor snapshot() {
        return cursor;
    }

    /** Empty when no build has ever started, or the persisted cursor belongs to a different
     * claimant (a fresh claim starts a fresh cursor — see {@link LiveBuildExecution}). */
    public Optional<BuildCursor> cursorFor(String claimantBody) {
        return cursor.claimantBody().equals(claimantBody) ? Optional.of(cursor) : Optional.empty();
    }

    /** Call after every {@link BuildEngine#advance} so a normal world save persists it —
     * mirrors {@code BoardData#markDirty}'s discipline. */
    public void save(BuildCursor cursor) {
        this.cursor = cursor;
        setDirty();
    }
}
