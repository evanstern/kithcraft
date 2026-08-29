package dev.kithcraft.mod.board;

import dev.kithcraft.mod.cast.Cast;
import net.minecraft.core.BlockPos;
import net.minecraft.core.component.DataComponents;
import net.minecraft.network.chat.Component;
import net.minecraft.server.level.ServerLevel;
import net.minecraft.world.item.ItemStack;
import net.minecraft.world.item.component.WritableBookContent;
import net.minecraft.world.item.component.WrittenBookContent;
import net.minecraft.world.level.block.Blocks;
import net.minecraft.world.level.block.entity.LecternBlockEntity;

import java.util.stream.Collectors;

/**
 * T001 (FR-001, card ACs #1, #7): the board's designated position. Plan.md's two options were
 * "a fixed config position" or "first-placed convention"; this picks the position — simpler,
 * and needs no block-place event hook (a new Fabric API registration this task does not need,
 * let alone a Mixin) to detect "the first lectern a player placed". The board sits inside
 * {@code dev.kithcraft.mod.brain.DuskPairing}'s own dusk-gathering arrival radius, so a
 * villager already walking there for {@code Activity.MEET} ({@link BoardVisit}) ends up
 * within reading range with no new AI.
 *
 * <p>Deliberately NOT registered as a vanilla POI (unlike {@code ScheduleSetup}'s job sites
 * and {@code DuskPairing}'s bell): a lectern is librarians' vanilla workstation, and this
 * board must never be claimable as one — skipping {@code PoiManager.add} keeps it invisible
 * to that unrelated vanilla system.
 *
 * <p>Untested by design, mirroring {@code ScheduleSetup}'s own precedent: thin {@code
 * ServerLevel}/{@code BlockEntity} glue with no branching logic worth a unit test; the
 * dev-server observation is its proof (Phase 4).
 */
public final class BoardSetup {
    private BoardSetup() {
    }

    /** Places the board's lectern at a fixed offset from {@code origin} (see class javadoc)
     * and returns its position. Re-placing the same block is a no-op in effect. */
    public static BlockPos placeBoard(ServerLevel level, BlockPos origin) {
        BlockPos pos = origin.offset(Cast.MEMBERS.size() + 2, 0, 2);
        level.setBlockAndUpdate(pos, Blocks.LECTERN.defaultBlockState());
        return pos;
    }

    /** Reads whatever book currently sits in the board's lectern into {@code board} — free
     * text, no parsing (card AC #7). No book present leaves the last-known posting untouched:
     * the artifact's content is what was posted, not what a prop happens to hold this tick.
     * ponytail: a real "book removed clears the posting" UX choice isn't asked for by any
     * Phase 1 card AC — Phase 3's build-snapshot-at-claim-time is what actually needs the
     * "book removed mid-build" edge case (spec edge cases), and reads the board fresh then. */
    public static void readBookInto(ServerLevel level, BlockPos boardPos, Board board) {
        if (!(level.getBlockEntity(boardPos) instanceof LecternBlockEntity lectern) || !lectern.hasBook()) {
            return;
        }
        String text = extractText(lectern.getBook());
        if (text != null) {
            board.post(text);
        }
    }

    /** A lectern accepts either an unsigned "book and quill" ({@code WRITABLE_BOOK_CONTENT})
     * or a signed written book ({@code WRITTEN_BOOK_CONTENT}) — card AC #7 needs neither form
     * nor syntax, so this reads whichever is present with no preference. */
    private static String extractText(ItemStack book) {
        WritableBookContent writable = book.get(DataComponents.WRITABLE_BOOK_CONTENT);
        if (writable != null) {
            return writable.getPages(false).collect(Collectors.joining("\n\n"));
        }
        WrittenBookContent written = book.get(DataComponents.WRITTEN_BOOK_CONTENT);
        if (written != null) {
            return written.getPages(false).stream()
                .map(Component::getString)
                .collect(Collectors.joining("\n\n"));
        }
        return null;
    }
}
