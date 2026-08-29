package dev.kithcraft.mod.board;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T002: the read cycle fires once per approach, never while stationary-near, and resets on
 * departure so a fresh visit fires again — mirrors PairingSignalTest's dedup coverage. */
class BoardReadTrackerTest {

    @Test
    void firesOnceOnArrivalThenNotAgainWhileStillNear() {
        BoardReadTracker tracker = new BoardReadTracker();
        assertTrue(tracker.shouldFireNow("b-tam", true), "first tick near fires");
        assertFalse(tracker.shouldFireNow("b-tam", true), "still near does not refire");
    }

    @Test
    void leavingClearsTheDedupSoAFreshVisitCanFireAgain() {
        BoardReadTracker tracker = new BoardReadTracker();
        assertTrue(tracker.shouldFireNow("b-tam", true));
        assertFalse(tracker.shouldFireNow("b-tam", false), "leaving never fires");
        assertTrue(tracker.shouldFireNow("b-tam", true), "a fresh approach fires again");
    }

    @Test
    void neverNearNeverFires() {
        BoardReadTracker tracker = new BoardReadTracker();
        assertFalse(tracker.shouldFireNow("b-tam", false));
    }

    @Test
    void tracksMultipleVillagersIndependently() {
        BoardReadTracker tracker = new BoardReadTracker();
        assertTrue(tracker.shouldFireNow("b-tam", true));
        assertTrue(tracker.shouldFireNow("b-petra", true), "a different villager's dedup is independent");
        assertFalse(tracker.shouldFireNow("b-tam", true));
    }
}
