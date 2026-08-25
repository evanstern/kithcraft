package dev.kithcraft.mod.live;

import dev.kithcraft.mod.act.Verbs;
import dev.kithcraft.mod.percept.Place;
import net.minecraft.core.BlockPos;
import net.minecraft.network.chat.Component;
import net.minecraft.world.entity.Mob;
import net.minecraft.world.entity.ai.navigation.PathNavigation;

import java.util.HashMap;
import java.util.Map;
import java.util.Objects;

/**
 * T011's minimal live {@link Verbs.Actuator}: a real villager's vanilla {@code
 * PathNavigation}/{@code LookControl}, plus a chat-adjacent broadcast for {@code speak} —
 * plain {@code Mob} API only, no Mixin (decision-0002; {@link Verbs}' class doc and R-8's
 * precedent both call for this). Symbols verified by {@code javap -public} against the
 * pinned MC 26.2 jar (specs/012-vendor-conformance/research/verb-observation.md).
 */
public final class LiveActuator implements Verbs.Actuator {
    private static final double WALK_SPEED = 0.5;
    private static final double ARRIVE_RADIUS = 1.5;

    private final Mob mob;
    private final LiveGround ground;
    // One body per actuator in practice (T011's single-body scope), but keyed defensively
    // rather than assumed — distinguishes "just started this walk" from "nav gave up".
    private final Map<String, BlockPos> activeWalks = new HashMap<>();

    public LiveActuator(Mob mob, LiveGround ground) {
        this.mob = Objects.requireNonNull(mob, "mob");
        this.ground = Objects.requireNonNull(ground, "ground");
    }

    @Override
    public Verbs.WalkStatus stepWalk(String body, Place place) {
        BlockPos target = ground.blockPosFor(place.place()).orElse(null);
        if (target == null) {
            activeWalks.remove(body);
            return Verbs.WalkStatus.TARGET_GONE;
        }
        if (mob.blockPosition().closerThan(target, ARRIVE_RADIUS)) {
            activeWalks.remove(body);
            mob.getNavigation().stop();
            return Verbs.WalkStatus.ARRIVED;
        }
        PathNavigation nav = mob.getNavigation();
        if (!target.equals(activeWalks.get(body))) {
            // A fresh walk toward this target: (re)issue the path.
            boolean started = nav.moveTo(target.getX() + 0.5, target.getY(), target.getZ() + 0.5, WALK_SPEED);
            if (!started) {
                activeWalks.remove(body);
                return Verbs.WalkStatus.UNREACHABLE;
            }
            activeWalks.put(body, target);
            return Verbs.WalkStatus.IN_PROGRESS;
        }
        if (nav.isDone()) {
            // Already walking this target; navigation ended without arriving.
            activeWalks.remove(body);
            return Verbs.WalkStatus.UNREACHABLE;
        }
        return Verbs.WalkStatus.IN_PROGRESS;
    }

    @Override
    public void speak(String body, Map<String, Object> target, String text) {
        if (text == null || mob.level().getServer() == null) {
            return;
        }
        // Chat-adjacent, not chat-real: villagers have no vanilla chat channel. A system
        // broadcast is the plain-API stand-in this seam's speak act drives (no Mixin).
        mob.level().getServer().getPlayerList().broadcastSystemMessage(Component.literal(text), false);
    }

    @Override
    public void attend(String body, Place place) {
        if (place == null) {
            return;
        }
        BlockPos target = ground.blockPosFor(place.place()).orElse(null);
        if (target != null) {
            mob.getLookControl().setLookAt(target.getX() + 0.5, target.getY() + 1.0, target.getZ() + 0.5);
        }
    }
}
