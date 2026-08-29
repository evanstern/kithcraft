package dev.kithcraft.mod.act;

import org.junit.jupiter.api.Test;

import java.util.Set;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T009: the declared verb set and the walk-status done() predicate IntentHandler drives on. */
class VerbsTest {

    @Test
    void declaredVerbsAreTheSection55CorePlusTask0020sClaimExtension() {
        assertEquals(Set.of("go_to", "speak", "attend", "wait", "claim"), Verbs.DECLARED);
    }

    @Test
    void onlyInProgressIsNotDone() {
        assertFalse(Verbs.WalkStatus.IN_PROGRESS.done());
        assertTrue(Verbs.WalkStatus.ARRIVED.done());
        assertTrue(Verbs.WalkStatus.UNREACHABLE.done());
        assertTrue(Verbs.WalkStatus.TARGET_GONE.done());
    }
}
