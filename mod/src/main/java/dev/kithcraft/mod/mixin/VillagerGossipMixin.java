package dev.kithcraft.mod.mixin;

import net.minecraft.server.level.ServerLevel;
import net.minecraft.world.entity.npc.villager.Villager;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfo;

/**
 * FR-004 override 2/2 — suppresses gossip exchange, and with it iron-golem summoning (card
 * AC #2). {@code specs/014-augmented-villager/research/brain-26.2.md} §5 traced
 * {@code Villager.gossip(ServerLevel, Villager, long)} calling {@code spawnGolemIfNeeded}
 * directly from inside its own body (bytecode: an {@code invokevirtual} to
 * {@code spawnGolemIfNeeded} at the tail of {@code gossip}'s method), and {@code gossip} is
 * invoked by {@code TradeWithVillager.tick()} — the same behavior class vanilla wires into
 * both the MEET and IDLE packages (confirmed by {@code javap -p -c} on
 * {@code VillagerGoalPackages} and {@code TradeWithVillager}, 2026-08-27). Cancelling
 * {@code gossip} at HEAD is therefore a single injection covering both suppression targets:
 * the Mixin budget for FR-004 (breeding + gossip + golem) lands at 2 overrides, not 3, per
 * the plan's own instruction to drop a symmetric-but-unneeded override.
 */
@Mixin(Villager.class)
public abstract class VillagerGossipMixin {

    @Inject(method = "gossip", at = @At("HEAD"), cancellable = true)
    private void kithcraft$suppressGossipAndGolem(ServerLevel level, Villager other, long gameTime, CallbackInfo ci) {
        ci.cancel();
    }
}
