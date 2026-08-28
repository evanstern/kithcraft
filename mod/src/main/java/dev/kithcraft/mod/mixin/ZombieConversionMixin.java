package dev.kithcraft.mod.mixin;

import net.minecraft.server.level.ServerLevel;
import net.minecraft.world.entity.monster.zombie.Zombie;
import net.minecraft.world.entity.npc.villager.Villager;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

/**
 * FR-003 (card AC #3) — makes zombie-villager conversion terminal-equivalent to death: no
 * cured path, no zombie-villager survivor identity (docs/design/death-mechanics.md §1 ruling).
 * {@code specs/019-death-remains/research/death-26.2.md}'s Phase-2 verification traced the
 * entry point: {@code Zombie.killedEntity(ServerLevel, LivingEntity, DamageSource)} decides,
 * on NORMAL (50% roll) or HARD (always) difficulty against a {@code Villager} victim, whether
 * to call {@code convertVillagerToZombieVillager(ServerLevel, Villager)} — a single
 * self-contained method whose entire body (bytecode offset 0) is
 * {@code Villager.convertTo(EntityTypes.ZOMBIE_VILLAGER, ...)}; nothing runs before that call.
 * Cancelling at HEAD to return {@code false} means {@code convertTo} never executes — no
 * entity substitution begins — and {@code killedEntity} falls through unaffected. Villager's
 * own {@code die(DamageSource)} (which released the victim's POIs synchronously, before
 * {@code super.die()} triggers this attacker-side hook) proceeds through its ordinary
 * loot-drop/removal path exactly as if the zombie had landed a normal killing blow, so Phase
 * 3's remains machinery fires for it like any other death — no extra routing code needed here.
 * One targeted injection, matching decision-0002's committed budget line (R1.3,
 * entity-implementation-comparison.md): V3's two (breeding, gossip) + siege (1) + this
 * conversion-cancel (1) = 4, the bound MixinConfigTest enforces.
 */
@Mixin(Zombie.class)
public abstract class ZombieConversionMixin {

    @Inject(method = "convertVillagerToZombieVillager", at = @At("HEAD"), cancellable = true)
    private void kithcraft$cancelConversion(ServerLevel level, Villager villager, CallbackInfoReturnable<Boolean> cir) {
        cir.setReturnValue(false);
    }
}
