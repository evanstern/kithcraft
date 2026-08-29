package dev.kithcraft.mod.build;

import dev.kithcraft.mod.cast.Cast;
import net.minecraft.core.BlockPos;

/**
 * T008: the build site's fixed position — the same "fixed offset from origin" idiom {@code
 * dev.kithcraft.mod.board.BoardSetup}/{@code dev.kithcraft.mod.brain.DuskPairing} already use,
 * for the identical reason {@code BoardSetup}'s own doc gives (no block-place/claim event
 * needed).
 *
 * <p>One global site, not one per villager: {@code dev.kithcraft.mod.board.Board} models
 * exactly one posting at a time (Phase 2's own finding), so at most one claim — and therefore
 * at most one build — is ever in progress. plan.md's "keep the single-posting model unless it
 * genuinely blocks" call, recorded here: it doesn't. A second concurrent build is out of this
 * demo's scope, not a corner cut around a real requirement.
 */
public final class BuildSetup {
    private BuildSetup() {
    }

    public static BlockPos siteOrigin(BlockPos origin) {
        return origin.offset(Cast.MEMBERS.size() + 2, 0, 5);
    }
}
