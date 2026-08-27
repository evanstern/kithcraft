package dev.kithcraft.mod.cast;

import dev.kithcraft.mod.brain.ScheduleSetup;
import net.minecraft.core.BlockPos;
import net.minecraft.core.Holder;
import net.minecraft.core.HolderLookup;
import net.minecraft.core.registries.BuiltInRegistries;
import net.minecraft.core.registries.Registries;
import net.minecraft.network.chat.Component;
import net.minecraft.resources.Identifier;
import net.minecraft.resources.ResourceKey;
import net.minecraft.server.level.ServerLevel;
import net.minecraft.world.entity.Entity;
import net.minecraft.world.entity.EntitySpawnReason;
import net.minecraft.world.entity.EntityType;
import net.minecraft.world.entity.npc.villager.Villager;
import net.minecraft.world.entity.npc.villager.VillagerData;
import net.minecraft.world.entity.npc.villager.VillagerProfession;
import net.minecraft.world.entity.npc.villager.VillagerType;

import java.util.List;

/**
 * Spawns {@link Cast#MEMBERS} into the world, once. Plain API only — no Mixin needed
 * (specs/014-augmented-villager/research/brain-26.2.md §6): a {@link VillagerData} sets
 * profession × biome variant, {@code Entity.setCustomName}/{@code setCustomNameVisible} gives
 * the nameplate. Idempotence and identity survival are {@link CastData}'s job; this class is
 * the thin, untested-by-design MC glue that only runs when that flag says seeding hasn't
 * happened yet — mirroring {@code dev.kithcraft.mod.live.BodySession}'s split between pure
 * core and live wiring, and {@code KithcraftMod}'s existing precedent of leaving live-server
 * glue to dev-server observation rather than a unit test.
 *
 * <p><b>Live finding (T006):</b> a code-spawned villager with a profession set directly (not
 * NONE), zero XP, and level 1 is exactly the state vanilla's own {@code ResetProfession} CORE
 * behavior treats as "claims a job but hasn't actually worked it" — confirmed by {@code javap
 * -p -c} on {@code net.minecraft.world.entity.ai.behavior.ResetProfession}: it fires once,
 * ~1 second after spawn, whenever profession is non-NONE/NITWIT, XP is 0, level ≤ 1, and no
 * {@code JOB_SITE} memory is held yet — and resets profession straight back to NONE. From
 * there, {@code AssignProfessionFromJobSite} reassigns the (now-unemployed) villager to
 * whichever unclaimed job site it happens to be nearest at that moment — not necessarily its
 * own, since the cast's three job-site blocks sit only a few blocks apart. Observed live as
 * two cast members' professions swapping. {@code Villager.setVillagerXp(int)} is public and
 * {@code ResetProfession}'s trigger short-circuits on nonzero XP, so seeding with 1 XP defuses
 * the reset at its actual condition — cleaner than placing job-site blocks far enough apart to
 * win the reassignment race, or duplicating vanilla's own POI-reservation bookkeeping by hand.
 */
public final class CastSeeder {
    private CastSeeder() {
    }

    private static final Identifier VILLAGER_ID = Identifier.fromNamespaceAndPath("minecraft", "villager");

    /** Spawns every not-yet-seeded cast member at {@code origin}, offset so they don't stack;
     * no-op if {@code data} says seeding already happened. */
    public static void seedIfNeeded(ServerLevel level, CastData data, BlockPos origin) {
        if (data.isSeeded()) {
            return;
        }
        EntityType<?> villagerType = BuiltInRegistries.ENTITY_TYPE.getValue(VILLAGER_ID);
        HolderLookup.RegistryLookup<VillagerProfession> professions =
            level.registryAccess().lookupOrThrow(Registries.VILLAGER_PROFESSION);
        HolderLookup.RegistryLookup<VillagerType> types =
            level.registryAccess().lookupOrThrow(Registries.VILLAGER_TYPE);

        var members = Cast.MEMBERS;
        forceLoadCastChunks(level, origin, members.size());
        for (int i = 0; i < members.size(); i++) {
            Cast.Member member = members.get(i);
            Entity entity = villagerType.create(level, EntitySpawnReason.COMMAND);
            if (!(entity instanceof Villager villager)) {
                continue;
            }
            Holder<VillagerProfession> profession = professions.getOrThrow(
                ResourceKey.create(Registries.VILLAGER_PROFESSION,
                    Identifier.fromNamespaceAndPath("minecraft", member.profession())));
            Holder<VillagerType> biomeVariant = types.getOrThrow(
                ResourceKey.create(Registries.VILLAGER_TYPE,
                    Identifier.fromNamespaceAndPath("minecraft", member.biomeVariant())));
            villager.setVillagerData(new VillagerData(biomeVariant, profession, 1));
            // See this class's own javadoc: defuses vanilla's ResetProfession behavior, which
            // would otherwise wipe this profession back to NONE about a second from now.
            villager.setVillagerXp(1);
            villager.setCustomName(Component.literal(member.name()));
            villager.setCustomNameVisible(true);
            BlockPos spawnPos = origin.offset(i * 2, 0, 0);
            villager.snapTo(spawnPos, 0f, 0f);
            level.addFreshEntity(villager);
            // T004 (FR-003): the bed/job-site fixtures vanilla scheduling needs to claim
            // something other than wander — see ScheduleSetup's own javadoc for why nothing
            // else about wake/work/socialize/sleep needs driving from here.
            ScheduleSetup.placeFixtures(level, villager, spawnPos, member);
        }
        data.markSeeded();
    }

    /** T004 finding: with zero players connected, MC 26.2's chunk-ticket system does not keep
     * any chunk loaded on its own (confirmed live — specs/014-augmented-villager/research/
     * schedule-observation.md), so the cast's own chunks unload within seconds and stop
     * ticking entirely (brains included) — the same "keep-loaded chunk tickets" gap
     * specs/012-vendor-conformance/research/verb-observation.md flagged as V3's to solve.
     * {@code ServerLevel.setChunkForced} is the public API behind {@code /forceload} — forced
     * chunks tick with no player nearby, which is exactly what a full unattended cycle (card
     * AC #1) needs. */
    private static void forceLoadCastChunks(ServerLevel level, BlockPos origin, int castSize) {
        int minChunkX = (origin.getX() - 2) >> 4;
        int maxChunkX = (origin.getX() + castSize * 2 + 2) >> 4;
        int minChunkZ = (origin.getZ() - 2) >> 4;
        int maxChunkZ = (origin.getZ() + 2) >> 4;
        for (int cx = minChunkX; cx <= maxChunkX; cx++) {
            for (int cz = minChunkZ; cz <= maxChunkZ; cz++) {
                level.setChunkForced(cx, cz, true);
            }
        }
    }

    /** Live finding, 2026-08-27 (specs/014-augmented-villager/research/
     * full-cycle-observation.md): a villager's own CORE-package wander can carry it beyond
     * {@link #forceLoadCastChunks}'s spawn-time footprint. A chunk one step past a forced
     * chunk still loads and ticks blocks/physics (confirmed via {@code javap} on {@code
     * ChunkLevel}: propagated ticket levels reach {@code BLOCK_TICKING} there, one level short
     * of the {@code ENTITY_TICKING} a forced chunk itself gets) but never runs Brain/goal-
     * selector AI — the villager reports live-looking NBT (small physics jitter in {@code
     * Motion}) forever, with position and Brain memories otherwise frozen, since nothing ever
     * ticks it into deciding to walk. {@code forceLoadCastChunks} alone can't self-heal this
     * (it only ever runs once, at first seed, per {@link #seedIfNeeded}'s idempotence guard) —
     * this method re-asserts a forced ticket around the cast's ACTUAL current positions, called
     * once per boot by {@code KithcraftMod} right after it finds the cast (the same
     * throttled-until-found idiom {@code DuskPairing.setUp}'s call site already uses), so a
     * restart always re-covers wherever the cast has actually wandered to, not just where it
     * spawned. */
    public static void keepChunksLoaded(ServerLevel level, List<Villager> castVillagers) {
        if (castVillagers.isEmpty()) {
            return;
        }
        int minChunkX = Integer.MAX_VALUE;
        int maxChunkX = Integer.MIN_VALUE;
        int minChunkZ = Integer.MAX_VALUE;
        int maxChunkZ = Integer.MIN_VALUE;
        for (Villager villager : castVillagers) {
            BlockPos pos = villager.blockPosition();
            minChunkX = Math.min(minChunkX, pos.getX() >> 4);
            maxChunkX = Math.max(maxChunkX, pos.getX() >> 4);
            minChunkZ = Math.min(minChunkZ, pos.getZ() >> 4);
            maxChunkZ = Math.max(maxChunkZ, pos.getZ() >> 4);
        }
        // Margin wide enough that CORE-package wander (observed live up to ~14 blocks from
        // spawn, schedule-observation.md) can never drift a villager past ENTITY_TICKING into
        // the degraded BLOCK_TICKING halo just beyond the forced set.
        int marginChunks = 3;
        for (int cx = minChunkX - marginChunks; cx <= maxChunkX + marginChunks; cx++) {
            for (int cz = minChunkZ - marginChunks; cz <= maxChunkZ + marginChunks; cz++) {
                level.setChunkForced(cx, cz, true);
            }
        }
    }
}
