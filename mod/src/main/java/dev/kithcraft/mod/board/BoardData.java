package dev.kithcraft.mod.board;

import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import net.minecraft.resources.Identifier;
import net.minecraft.util.datafix.DataFixTypes;
import net.minecraft.world.level.saveddata.SavedData;
import net.minecraft.world.level.saveddata.SavedDataType;

/**
 * Persists {@link Board} through the world save (T001), the same adapter pattern {@code
 * dev.kithcraft.mod.tokens.TokenRegistryData}/{@code dev.kithcraft.mod.cast.CastData}
 * established: a codec-described {@link SavedDataType}, the pure {@link Board} owning every
 * rule, this class only getting it onto disk.
 */
public final class BoardData extends SavedData {

    private static final Codec<Board.Snapshot> SNAPSHOT_CODEC = RecordCodecBuilder.create(i -> i.group(
        Codec.STRING.fieldOf("text").forGetter(Board.Snapshot::text),
        Codec.STRING.listOf().fieldOf("claims").forGetter(Board.Snapshot::claims)
    ).apply(i, Board.Snapshot::new));

    public static final Codec<BoardData> CODEC = SNAPSHOT_CODEC.xmap(BoardData::new, BoardData::snapshot);

    public static final SavedDataType<BoardData> TYPE = new SavedDataType<>(
        Identifier.fromNamespaceAndPath("kithcraft", "board"),
        BoardData::new, CODEC, DataFixTypes.LEVEL);

    private final Board board;

    public BoardData() {
        this.board = new Board();
    }

    private BoardData(Board.Snapshot snapshot) {
        this.board = Board.restore(snapshot);
    }

    private Board.Snapshot snapshot() {
        return board.snapshot();
    }

    public Board board() {
        return board;
    }

    /** Call after any mutation of {@link #board()} (posting, claims) so a normal world save
     * persists it — mirrors {@code TokenRegistryData}'s dirty-flag discipline. */
    public void markDirty() {
        setDirty();
    }
}
