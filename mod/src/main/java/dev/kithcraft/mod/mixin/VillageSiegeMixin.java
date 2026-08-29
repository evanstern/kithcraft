package dev.kithcraft.mod.mixin;

import net.minecraft.server.level.ServerLevel;
import net.minecraft.world.entity.ai.village.VillageSiege;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfo;

/**
 * FR-002 (card AC #2) — suppresses zombie sieges unconditionally, regardless of village
 * eligibility. {@code specs/019-death-remains/research/death-26.2.md} R-5 traced
 * {@code VillageSiege.tick(ServerLevel, boolean)} as the sole per-tick entry point of the
 * entire siege mechanism ({@code VillageSiege implements CustomSpawner}, one class, no state
 * to partially unwind): a clock-marker gate, a 10% nightly roll, eligibility/setup, then a
 * 20-zombie spawn drip all run from calls made inside this one method. Cancelling at HEAD stops
 * the state machine before any of that runs — no roll, no setup, no spawn — same shape as the
 * landed {@link VillagerGossipMixin} (single {@code @Inject(HEAD, cancellable)}). Per
 * death-26.2.md R-5b, eligibility is pure POI-occupancy density with no door/population count
 * in 26.2, so a 3-villager cast always qualifies once any bed/job-site POI is claimed — the
 * plan's design decision 2 ("suppress regardless of eligibility") is real work here, not
 * covering an unreachable case.
 */
@Mixin(VillageSiege.class)
public abstract class VillageSiegeMixin {

    @Inject(method = "tick", at = @At("HEAD"), cancellable = true)
    private void kithcraft$suppressSiege(ServerLevel level, boolean spawnHostiles, CallbackInfo ci) {
        ci.cancel();
    }
}
