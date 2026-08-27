package dev.kithcraft.mod.cast;

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

/**
 * Spawns {@link Cast#MEMBERS} into the world, once. Plain API only — no Mixin needed
 * (specs/014-augmented-villager/research/brain-26.2.md §6): a {@link VillagerData} sets
 * profession × biome variant, {@code Entity.setCustomName}/{@code setCustomNameVisible} gives
 * the nameplate. Idempotence and identity survival are {@link CastData}'s job; this class is
 * the thin, untested-by-design MC glue that only runs when that flag says seeding hasn't
 * happened yet — mirroring {@code dev.kithcraft.mod.live.BodySession}'s split between pure
 * core and live wiring, and {@code KithcraftMod}'s existing precedent of leaving live-server
 * glue to dev-server observation rather than a unit test.
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
            villager.setCustomName(Component.literal(member.name()));
            villager.setCustomNameVisible(true);
            villager.snapTo(origin.offset(i * 2, 0, 0), 0f, 0f);
            level.addFreshEntity(villager);
        }
        data.markSeeded();
    }
}
