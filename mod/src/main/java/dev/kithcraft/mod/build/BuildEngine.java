package dev.kithcraft.mod.build;

import java.util.List;

/**
 * T008/T009 (spec.md US3/US4, card ACs #5, #6, #8): steps a claimed build forward by one
 * tick's worth of placement, gated on a single boolean the caller computes however it likes —
 * real wiring passes {@code villager.getBrain().isActive(Activity.WORK)} ({@link
 * LiveBuildExecution}). This class does not know or care WHY {@code working} went false — a
 * dusk schedule transition and a danger-driven PANIC both collapse to the same boolean
 * (villager-brain-api.md's addActivity-additive facts: PANIC simply outranks WORK on the same
 * vanilla schedule substrate), which is exactly why both interrupt kinds need no separate code
 * path here, only separate tests proving the one mechanism covers both ({@code
 * BuildEngineTest}).
 *
 * <p>Tick budget ({@link #BLOCKS_PER_TICK}) is deliberately small — "beside the player," not a
 * flash-build (kithcraft-brief.md's loneliness-cure constraint). No re-issue/supervise path
 * exists: the only inputs are the persisted cursor, the plan, and one boolean; nothing here
 * accepts player-supplied materials or forced progress (card AC #8) — {@code
 * StructuralAbsenceTest#noManualBuildProgressPath} checks this method has exactly one call
 * site in the whole mod.
 */
public final class BuildEngine {
    private BuildEngine() {
    }

    /** ponytail: one block/tick is a deliberate pace choice for the demo, not a tuned budget —
     * ceiling is "visibly incremental, not instant"; upgrade path is a config value (mirrors
     * {@code GriefPeriod.configuredTicks}'s system-property idiom) if the pace ever needs
     * retuning — not reached for here since nothing yet asks to. */
    public static final int BLOCKS_PER_TICK = 1;

    public record Step(BuildCursor cursor, List<Placement.Offset> toPlace, boolean done) {
    }

    /** {@code working=false} (interrupted) places nothing and returns {@code cursor}
     * unchanged — the persisted cursor IS the pause, nothing else is needed. {@code
     * working=true} places up to {@link #BLOCKS_PER_TICK} more offsets from {@code plan},
     * starting exactly where {@code cursor} left off. A cursor already at or past the plan's
     * end is idempotent: no further placement, {@code done} stays true. */
    public static Step advance(BuildCursor cursor, List<Placement.Offset> plan, boolean working) {
        boolean alreadyDone = cursor.placed() >= plan.size();
        if (!working || alreadyDone) {
            return new Step(cursor, List.of(), alreadyDone);
        }
        int end = Math.min(cursor.placed() + BLOCKS_PER_TICK, plan.size());
        List<Placement.Offset> toPlace = plan.subList(cursor.placed(), end);
        BuildCursor next = new BuildCursor(cursor.claimantBody(), cursor.blueprintText(), end);
        return new Step(next, toPlace, end >= plan.size());
    }
}
