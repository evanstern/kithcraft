package dev.kithcraft.mod.mixin;

import com.google.common.collect.ImmutableList;
import com.mojang.datafixers.util.Pair;
import net.minecraft.world.entity.ai.behavior.BehaviorControl;
import net.minecraft.world.entity.ai.behavior.VillagerGoalPackages;
import net.minecraft.world.entity.ai.behavior.VillagerMakeLove;
import net.minecraft.world.entity.npc.villager.Villager;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

/**
 * FR-004 override 1/2 — suppresses villager breeding (card AC #2).
 *
 * <p>Corrects a finding in {@code specs/014-augmented-villager/research/brain-26.2.md} §5,
 * which (by class-existence search only) placed {@link VillagerMakeLove} in the CORE package.
 * Re-verified here by {@code javap -p -c} against {@code VillagerGoalPackages} (2026-08-27):
 * {@code VillagerMakeLove} is actually constructed inside {@code getIdlePackage(float)}, not
 * {@code getCorePackage(Holder, float)} — bytecode offset for {@code new VillagerMakeLove}
 * falls strictly between {@code getIdlePackage}'s method header and the next method
 * ({@code getPanicPackage}), never inside {@code getCorePackage}'s range. This mixin targets
 * the actually-correct method.
 *
 * <p>Villager's own {@code registerBrainGoals} does not call {@code addActivity} at all in
 * 26.2 (it only sets the schedule attribute) — the real activity-package wiring runs through
 * the static {@code Brain.ActivitySupplier} lambda baked into {@code Villager.BRAIN_PROVIDER},
 * which calls {@code VillagerGoalPackages.getIdlePackage(float)} directly and feeds the result
 * straight into the {@code Brain} constructor, never through the public (additive-only)
 * {@code Brain.addActivity}. So the only place to omit a stock behavior is the source package
 * method itself, filtered here before any {@code Brain} is built from it.
 */
@Mixin(VillagerGoalPackages.class)
public abstract class VillagerGoalPackagesMixin {

    @Inject(method = "getIdlePackage", at = @At("RETURN"), cancellable = true)
    private static void kithcraft$suppressBreeding(
            float speedModifier,
            CallbackInfoReturnable<ImmutableList<Pair<Integer, ? extends BehaviorControl<? super Villager>>>> cir) {
        ImmutableList<Pair<Integer, ? extends BehaviorControl<? super Villager>>> withoutBreeding =
            cir.getReturnValue().stream()
                .filter(pair -> !(pair.getSecond() instanceof VillagerMakeLove))
                .collect(ImmutableList.toImmutableList());
        cir.setReturnValue(withoutBreeding);
    }
}
