package dev.kithcraft.mod.build;

import dev.kithcraft.mod.board.Board;
import net.minecraft.core.BlockPos;
import net.minecraft.server.level.ServerLevel;
import net.minecraft.world.entity.npc.villager.Villager;
import net.minecraft.world.entity.schedule.Activity;
import net.minecraft.world.level.block.Blocks;
import net.minecraft.world.level.block.state.BlockState;

import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.function.Function;

/**
 * T008/T009 (spec.md US3/US4, card ACs #5, #6, #8): the build's engine-side placement, driven
 * every server tick. Thin {@code ServerLevel}/{@code Brain} glue only — every decision (parse,
 * plan, step, resume) lives in {@link BlueprintParser}/{@link Placement}/{@link BuildEngine},
 * all pure and unit tested; this class only (a) finds the claimed villager, (b) reads whether
 * it is currently {@code Activity.WORK}-active, and (c) applies whatever offsets {@link
 * BuildEngine} hands back to the real world. No player-facing entry point exists here or
 * anywhere else to supply materials or force progress — the only driver is this tick method,
 * gated on the villager's own vanilla schedule ({@code StructuralAbsenceTest
 * #noManualBuildProgressPath}).
 *
 * <p><b>Material sourcing (plan.md design decision 5, card AC #8's no-player-supply-path
 * requirement):</b> the simplest defensible rule — creative-style provision. A placed block
 * simply appears; nothing is drawn from a persisted inventory. ponytail: real village-storage
 * withdrawal (a designated chest, decremented per block) is the honest upgrade path if the
 * demo ever wants materials to feel scarce; deliberately out of scope for "the thinnest
 * possible build system" (kithcraft-brief.md).
 *
 * <p>ponytail: spec Edge Cases' "unreachable build site: claim stands, build attempts pause
 * and retry" does not apply to this design — placement is a direct world write, not a
 * villager path to each block, so there is no pathing failure mode to pause/retry against.
 * Recorded honestly as an edge case this engine-side-instant-placement design structurally
 * sidesteps, not one it handles.
 */
public final class LiveBuildExecution {
    private final BlockPos siteOrigin;
    private final Board board;
    private final BuildData buildData;

    public LiveBuildExecution(BlockPos siteOrigin, Board board, BuildData buildData) {
        this.siteOrigin = siteOrigin;
        this.board = board;
        this.buildData = buildData;
    }

    /** Called once per server tick. {@code bodyTokenLookup} resolves a villager's own body
     * token, the same lookup {@code dev.kithcraft.mod.board.BoardVisit} already uses. */
    public void tick(ServerLevel level, List<Villager> villagers, Function<UUID, Optional<String>> bodyTokenLookup) {
        Optional<String> claimant = board.claimedBy();
        if (claimant.isEmpty()) {
            return;
        }
        String claimantBody = claimant.get();
        Villager villager = findClaimant(villagers, bodyTokenLookup, claimantBody);
        if (villager == null) {
            return; // claimant not currently loaded/found this tick — try again next tick
        }
        BuildCursor existing = buildData.cursorFor(claimantBody).orElse(null);
        // Card AC #6/Edge Cases: the blueprint text is captured once, at start — a later board
        // edit never reaches an in-progress build; only a brand-new cursor reads the live text.
        String blueprintText = existing != null ? existing.blueprintText() : board.text();
        Optional<Blueprint> blueprint = BlueprintParser.parse(blueprintText);
        if (blueprint.isEmpty()) {
            return; // unparseable posting (no text at all) — nothing to build
        }
        BuildCursor cursor = existing != null ? existing : BuildCursor.start(claimantBody, blueprintText);
        List<Placement.Offset> plan = Placement.plan(blueprint.get());
        boolean working = villager.getBrain().isActive(Activity.WORK);
        BuildEngine.Step step = BuildEngine.advance(cursor, plan, working);
        for (Placement.Offset offset : step.toPlace()) {
            level.setBlockAndUpdate(siteOrigin.offset(offset.x(), offset.y(), offset.z()), blockStateFor(blueprint.get().material()));
        }
        buildData.save(step.cursor());
    }

    private static Villager findClaimant(
            List<Villager> villagers, Function<UUID, Optional<String>> bodyTokenLookup, String claimantBody) {
        return villagers.stream()
            .filter(v -> bodyTokenLookup.apply(v.getUUID()).map(claimantBody::equals).orElse(false))
            .findFirst().orElse(null);
    }

    private static BlockState blockStateFor(Blueprint.Material material) {
        return switch (material) {
            case WOOD -> Blocks.OAK_PLANKS.defaultBlockState();
            case STONE -> Blocks.COBBLESTONE.defaultBlockState();
            case BRICK -> Blocks.BRICKS.defaultBlockState();
            case DIRT -> Blocks.DIRT.defaultBlockState();
        };
    }
}
