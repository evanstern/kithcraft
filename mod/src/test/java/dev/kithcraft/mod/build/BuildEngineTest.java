package dev.kithcraft.mod.build;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T008/T009 (spec.md US3/US4, card ACs #5, #6, #8): tick-budget placement stepping and
 * interrupt/resume. {@link BuildEngine#advance} treats "not working" as the single interrupt
 * signal — real wiring computes it from {@code Activity.WORK}, which goes false for BOTH a
 * dusk schedule transition and a danger-driven PANIC (same vanilla schedule substrate, PANIC
 * simply outranks WORK). The two interrupt tests below are deliberately mechanically
 * identical — that IS the point: one mechanism covers both interrupt kinds card AC #6 names,
 * not two. */
class BuildEngineTest {

    private static final List<Placement.Offset> PLAN =
        Placement.plan(new Blueprint(Blueprint.Shape.TOWER, 5, Blueprint.Material.STONE));

    @Test
    void placesOneBlockPerTickWhileWorking() {
        BuildCursor cursor = BuildCursor.start("b-tam", "a stone tower");
        BuildEngine.Step step1 = BuildEngine.advance(cursor, PLAN, true);
        assertEquals(1, step1.toPlace().size());
        assertEquals(new Placement.Offset(0, 0, 0), step1.toPlace().get(0));
        assertEquals(1, step1.cursor().placed());

        BuildEngine.Step step2 = BuildEngine.advance(step1.cursor(), PLAN, true);
        assertEquals(new Placement.Offset(0, 1, 0), step2.toPlace().get(0));
        assertEquals(2, step2.cursor().placed());
    }

    @Test
    void scheduleTransitionInterruptsPlacementAndCursorPersistsUnchanged() {
        BuildCursor cursor = BuildEngine.advance(BuildCursor.start("b-tam", "a stone tower"), PLAN, true).cursor();
        BuildEngine.Step interrupted = BuildEngine.advance(cursor, PLAN, false); // dusk: no longer WORK-active
        assertTrue(interrupted.toPlace().isEmpty(), "nothing places while interrupted");
        assertEquals(cursor, interrupted.cursor(), "the cursor is exactly the pause — nothing else needed");
    }

    @Test
    void dangerPanicInterruptsPlacementAndCursorPersistsUnchanged() {
        BuildCursor cursor = BuildEngine.advance(BuildCursor.start("b-tam", "a stone tower"), PLAN, true).cursor();
        BuildEngine.Step interrupted = BuildEngine.advance(cursor, PLAN, false); // panic: no longer WORK-active either
        assertTrue(interrupted.toPlace().isEmpty());
        assertEquals(cursor, interrupted.cursor());
    }

    @Test
    void resumesFromTheCursorWithoutRestartingOrReclaiming() {
        BuildCursor afterOneBlock = BuildEngine.advance(BuildCursor.start("b-tam", "a stone tower"), PLAN, true).cursor();
        BuildCursor stillPausedAtNextTick = BuildEngine.advance(afterOneBlock, PLAN, false).cursor();
        // "Resume" is just calling advance() again on the SAME persisted cursor with
        // working=true — no claim-related argument exists to pass here at all.
        BuildEngine.Step resumed = BuildEngine.advance(stillPausedAtNextTick, PLAN, true);
        assertEquals(new Placement.Offset(0, 1, 0), resumed.toPlace().get(0), "picks up at block 2, not block 1 again");
        assertEquals(2, resumed.cursor().placed());
    }

    @Test
    void completionIsIdempotentNoOverbuildOnRepeatedTicks() {
        BuildCursor cursor = new BuildCursor("b-tam", "a stone tower", PLAN.size());
        BuildEngine.Step step = BuildEngine.advance(cursor, PLAN, true);
        assertTrue(step.done());
        assertTrue(step.toPlace().isEmpty());
        assertEquals(cursor, step.cursor(), "a finished build advances no further");
    }
}
