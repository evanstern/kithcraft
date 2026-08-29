package dev.kithcraft.mod.death;

/**
 * R-6 (death-mechanics.md §6.2 item 5, plan.md design decision 3): a danger-tuning knob for a
 * scripted/recorded demo take that cannot afford to lose a cast member mid-take. OFF by
 * default — this is an explicit per-run operator choice, never the default posture. It does
 * NOT touch permadeath or the admitted-causes table (§5/§6.2 are explicit that the fix for a
 * losable take is thinning the environmental danger, not changing the rules); same
 * system-property idiom as {@link GriefPeriod}.
 *
 * <p>ponytail: the smallest defensible lever — despawn a newly-spawned hostile that lands
 * within {@link #SUPPRESSION_RADIUS} blocks of a cast member, wired at {@code KithcraftMod}'s
 * already-registered {@code ServerEntityEvents.ENTITY_LOAD} hook (no new Mixin, no new event
 * registration — FR-006's 4-Mixin cap stays spent exactly where it is). A real
 * hostile-spawn-pressure scale would need a Mixin on {@code NaturalSpawner}; that's the
 * upgrade path if a recorded take needs more than "nothing hostile spawns adjacent to the
 * cast," spent deliberately against the Mixin budget rather than smuggled in here.
 */
public final class DangerTuning {
    private DangerTuning() {}

    public static final double SUPPRESSION_RADIUS = 24.0;

    /** Config, not a constant (card AC #4/#6): off unless explicitly opted into for a take. */
    public static boolean enabled() {
        return Boolean.getBoolean("kithcraft.dangerTuning");
    }

    /** Pure decision, no Minecraft types: suppress iff tuning is on and the hostile landed
     * within radius of a cast member. Pass {@code Double.POSITIVE_INFINITY} when no cast
     * member is loaded. */
    public static boolean shouldSuppress(boolean enabled, double distanceToNearestCastMember) {
        return enabled && distanceToNearestCastMember <= SUPPRESSION_RADIUS;
    }
}
