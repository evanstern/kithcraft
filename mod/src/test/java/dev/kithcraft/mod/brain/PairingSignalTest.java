package dev.kithcraft.mod.brain;

import org.junit.jupiter.api.Test;

import java.util.Map;
import java.util.OptionalDouble;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T008: the pairing signal's timing math (fires on predicted-arrival-minus-~10s, never on
 * arrival) and the no-fire path (pathing failure). */
class PairingSignalTest {

    @Test
    void predictedArrivalIsRemainingDistanceOverSpeed() {
        assertEquals(10.0, PairingSignal.predictedArrivalSeconds(OptionalDouble.of(100), 10).getAsDouble());
    }

    @Test
    void noFireWhenPathingHasFailed() {
        assertTrue(PairingSignal.predictedArrivalSeconds(OptionalDouble.empty(), 10).isEmpty());
    }

    @Test
    void noFireWhenStuckWithNonPositiveSpeed() {
        assertTrue(PairingSignal.predictedArrivalSeconds(OptionalDouble.of(100), 0).isEmpty());
    }

    @Test
    void combineTakesTheSlowerMembersEta() {
        assertEquals(15.0,
            PairingSignal.combine(OptionalDouble.of(5.0), OptionalDouble.of(15.0)).getAsDouble());
    }

    @Test
    void combineIsEmptyIfEitherMemberFailedToPath() {
        assertTrue(PairingSignal.combine(OptionalDouble.of(5.0), OptionalDouble.empty()).isEmpty());
    }

    @Test
    void firesOnceAtTheLeadWindowBoundaryThenNotAgain() {
        PairingSignal signal = new PairingSignal();
        assertTrue(signal.shouldFireNow("a+b", OptionalDouble.of(PairingSignal.LEAD_SECONDS)),
            "first tick inside the lead window fires");
        assertFalse(signal.shouldFireNow("a+b", OptionalDouble.of(9.0)),
            "a later tick of the same approach does not refire");
    }

    @Test
    void doesNotFireBeforeTheLeadWindow() {
        PairingSignal signal = new PairingSignal();
        assertFalse(signal.shouldFireNow("a+b", OptionalDouble.of(15.0)), "15s out is not yet ~10s out");
    }

    @Test
    void doesNotFireOnOrAfterArrival() {
        PairingSignal signal = new PairingSignal();
        assertFalse(signal.shouldFireNow("a+b", OptionalDouble.of(0.0)), "arrival itself must not fire (R-7)");
    }

    @Test
    void noFireWhenTheCombinedEtaIsEmpty() {
        PairingSignal signal = new PairingSignal();
        assertFalse(signal.shouldFireNow("a+b", OptionalDouble.empty()), "pathing failure never fires");
    }

    @Test
    void resetAllowsAFreshApproachToFireAgain() {
        PairingSignal signal = new PairingSignal();
        assertTrue(signal.shouldFireNow("a+b", OptionalDouble.of(5.0)));
        signal.reset("a+b");
        assertTrue(signal.shouldFireNow("a+b", OptionalDouble.of(5.0)), "a reset approach can fire again");
    }

    @Test
    void contentIsASightingOfTheOtherMemberCarryingTheirBodyToken() {
        Map<String, Object> content = PairingSignal.content("b-petra", "Petra");
        assertEquals("near", content.get("distance"));
        Map<?, ?> thing = (Map<?, ?>) content.get("thing");
        assertEquals("k:person", thing.get("kind"));
        assertEquals("b-petra", thing.get("body"));
        assertEquals("Petra", thing.get("descriptor"));
    }
}
