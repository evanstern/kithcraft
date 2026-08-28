package dev.kithcraft.mod.brain;

import dev.kithcraft.mod.cast.Cast;
import net.minecraft.core.BlockPos;
import net.minecraft.core.Direction;
import net.minecraft.core.GlobalPos;
import net.minecraft.server.level.ServerLevel;
import net.minecraft.world.entity.ai.memory.MemoryModuleType;
import net.minecraft.world.entity.ai.village.poi.PoiTypes;
import net.minecraft.world.entity.npc.villager.Villager;
import net.minecraft.world.level.block.BedBlock;
import net.minecraft.world.level.block.Block;
import net.minecraft.world.level.block.Blocks;
import net.minecraft.world.level.block.HorizontalDirectionalBlock;
import net.minecraft.world.level.block.state.BlockState;
import net.minecraft.world.level.block.state.properties.BedPart;

import java.util.Map;

/**
 * T004 (FR-003): places the one fixture vanilla scheduling needs that {@link
 * dev.kithcraft.mod.cast.CastSeeder} does not already provide — a claimable bed and a
 * claimable job-site block next to each cast member. Wake/work/socialize/sleep, POI
 * claim/release, sleep pathing, and door use are all vanilla's own Brain/Sensor/Behavior
 * machinery, driven automatically once a profession-bearing {@code Villager} exists and a
 * matching POI is in range (specs/014-augmented-villager/research/brain-26.2.md §§1,3,5) —
 * this class adds no scheduling logic of its own, only the world blocks vanilla's own sensors
 * already look for. It is safe to cluster these fixtures a few blocks apart (rather than
 * needing a real, spread-out village) precisely because vanilla's own claim behaviors filter
 * by profession — see {@link dev.kithcraft.mod.cast.CastSeeder}'s javadoc for the live finding
 * that made that filtering load-bearing (T006).
 *
 * <p>The profession → job-site block mapping is vanilla's own long-standing, data-driven
 * poi_type registration (unchanged since Village &amp; Pillage), not something this task's
 * javap passes derive — it lives in datapack JSON, not code. {@link Cast#MEMBERS}' three
 * professions: armorer/blast furnace, farmer/composter, fisherman/barrel.
 *
 * <p>ponytail: one bed + one job site per member, placed in open air with no walls or doors —
 * door-use (FR-003) is vanilla behaviour this doesn't need to exercise to prove wake/work/
 * sleep; a real village structure is cosmetic polish, out of this task's scope.
 *
 * <p><b>Live finding (T006), two layers deep:</b> placing the job-site block is not enough on
 * its own to give an already-professioned villager a {@code JOB_SITE} memory, for two
 * independent reasons found live, in order.
 * <ol>
 *   <li>Vanilla only ever assigns {@code JOB_SITE} as part of {@code
 *   AssignProfessionFromJobSite}'s NONE-profession-to-employed transition (the bed's
 *   {@code HOME} memory self-claims fine via a separate, profession-independent CORE
 *   behavior, but {@code JOB_SITE} never did — even after 30+ real seconds standing directly
 *   on the block). Since {@link dev.kithcraft.mod.cast.CastSeeder} assigns the profession
 *   directly and deliberately never routes through that transition, this claims
 *   {@code JOB_SITE} directly on the {@code Brain} instead.</li>
 *   <li>Setting the Brain memory alone was still not enough: it kept disappearing.
 *   {@code ValidateNearbyPoi} (CORE package, confirmed by {@code javap -p -c}) erases
 *   {@code JOB_SITE} the moment {@code PoiManager.exists(pos, matchingType)} comes back false
 *   — and a block placed via {@code setBlockAndUpdate} during {@code onServerStarted} is not
 *   guaranteed to have already reached the world's {@link
 *   net.minecraft.world.entity.ai.village.poi.PoiManager} by the time that check first runs.
 *   {@code PoiTypes.forState} plus an explicit {@code PoiManager.add} makes the registration
 *   synchronous with the block placement, closing that race.</li>
 * </ol>
 *
 * <p>Untested by design, mirroring {@link dev.kithcraft.mod.cast.CastSeeder}'s own precedent:
 * this is thin {@code ServerLevel}/{@code Brain} glue with no branching logic worth a unit
 * test; the dev-server observation (T006) is its proof.
 */
public final class ScheduleSetup {
    private ScheduleSetup() {
    }

    private static final Map<String, Block> JOB_SITE_BLOCKS = Map.of(
        "armorer", Blocks.BLAST_FURNACE,
        "farmer", Blocks.COMPOSTER,
        "fisherman", Blocks.BARREL
    );

    /** Places a claimable bed and (if the member's profession has one) a claimable job-site
     * block beside {@code villagerPos}, and claims the job site directly on {@code villager}'s
     * Brain (see class javadoc for why placing alone isn't enough). Re-placing the same
     * blocks/memory is a no-op in effect; call-site idempotence (don't call this twice) is
     * {@link dev.kithcraft.mod.cast.CastData}'s job, same as the villager spawn itself. */
    public static void placeFixtures(ServerLevel level, Villager villager, BlockPos villagerPos, Cast.Member member) {
        Block jobSite = JOB_SITE_BLOCKS.get(member.profession());
        if (jobSite != null) {
            BlockPos jobSitePos = villagerPos.offset(1, 0, 1);
            BlockState jobSiteState = jobSite.defaultBlockState();
            level.setBlockAndUpdate(jobSitePos, jobSiteState);
            if (level.getPoiManager().getType(jobSitePos).isEmpty()) {
                PoiTypes.forState(jobSiteState).ifPresent(poiType -> level.getPoiManager().add(jobSitePos, poiType));
            }
            villager.getBrain().setMemory(MemoryModuleType.JOB_SITE, GlobalPos.of(level.dimension(), jobSitePos));
        }

        BlockState bed = Blocks.BED.red().defaultBlockState()
            .setValue(HorizontalDirectionalBlock.FACING, Direction.NORTH);
        BlockPos foot = villagerPos.offset(-1, 0, -1);
        BlockPos head = foot.relative(Direction.NORTH);
        level.setBlockAndUpdate(foot, bed.setValue(BedBlock.PART, BedPart.FOOT));
        level.setBlockAndUpdate(head, bed.setValue(BedBlock.PART, BedPart.HEAD));
    }
}
