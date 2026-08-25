package dev.kithcraft.mod.act;

import org.junit.jupiter.api.Test;

import java.util.Set;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T009: the declared verb set and the walk-status done() predicate IntentHandler drives on. */
class VerbsTest {

    @Test
    void declaredVerbsAreExactlyTheSection55Core() {
        assertEquals(Set.of("go_to", "speak", "attend", "wait"), Verbs.DECLARED);
    }

    @Test
    void onlyInProgressIsNotDone() {
        assertFalse(Verbs.WalkStatus.IN_PROGRESS.done());
        assertTrue(Verbs.WalkStatus.ARRIVED.done());
        assertTrue(Verbs.WalkStatus.UNREACHABLE.done());
        assertTrue(Verbs.WalkStatus.TARGET_GONE.done());
    }
}
