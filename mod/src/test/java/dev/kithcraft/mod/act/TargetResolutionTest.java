package dev.kithcraft.mod.act;

import dev.kithcraft.mod.percept.Place;
import org.junit.jupiter.api.Test;

import java.util.Map;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T008 (card AC #7/#8, §5.6): the three named target-resolution cases. */
class TargetResolutionTest {

    private static TargetResolution.Ground stubGround(String issuedToken, Place place) {
        return new TargetResolution.Ground() {
            @Override public boolean isIssued(String token) {
                return token.equals(issuedToken);
            }
            @Override public Optional<Place> placeFor(String token) {
                return token.equals(issuedToken) ? Optional.ofNullable(place) : Optional.empty();
            }
        };
    }

    /** Card AC #8's first half: an intent naming a token the vendor never issued is refused
     * unknown_target — the only legitimate case for that refusal (§5.3). */
    @Test
    void unissuedTokenIsRefusedUnknownTarget() {
        TargetResolution.Ground ground = stubGround("b-tam", new Place("pl-1", "the well"));
        TargetResolution.Result result = TargetResolution.resolve(ground, Map.of("type", "body", "body", "b-ghost"));
        assertEquals(TargetResolution.Outcome.UNKNOWN_TARGET, result.outcome());
    }

    /** Card AC #7: a body target resolves to that body's LAST-SEEN place. {@link
     * TargetResolution.Ground} exposes no live-position lookup at all — {@code placeFor} is
     * the only place a body target can come from, so "never live position" holds by
     * construction, not by a runtime check that could be bypassed. */
    @Test
    void bodyTargetResolvesToTheLastSeenPlaceOnly() {
        Place lastSeen = new Place("pl-51c", "the old orchard");
        TargetResolution.Ground ground = stubGround("b-tam", lastSeen);
        TargetResolution.Result result = TargetResolution.resolve(ground, Map.of("type", "body", "body", "b-tam"));
        assertEquals(TargetResolution.Outcome.RESOLVED, result.outcome());
        assertEquals(lastSeen, result.place());
    }

    /** Card AC #8's second half: a token the vendor DID issue, whose referent has since
     * become unreachable/gone, is ACCEPTED here regardless — resolution has no existence
     * check to fail. {@link IntentHandlerTest} proves the failure surfaces only after the
     * walk actually tries (§5.6). */
    @Test
    void aKnownButNowGoneReferentStillResolvesAcceptedNotRefused() {
        Place lastSeen = new Place("pl-51c", "the old orchard");
        TargetResolution.Ground ground = stubGround("b-tam", lastSeen);
        // "gone" is not a fact TargetResolution has any way to represent — the same issued
        // token, with the same last-seen place, resolves identically whether or not the
        // referent still exists. That is the point: no live check happens here.
        TargetResolution.Result result = TargetResolution.resolve(ground, Map.of("type", "body", "body", "b-tam"));
        assertEquals(TargetResolution.Outcome.RESOLVED, result.outcome());
        assertTrue(ground.isIssued("b-tam"));
    }

    @Test
    void placeTargetResolvesDirectly() {
        Place well = new Place("pl-1", "the well");
        TargetResolution.Ground ground = stubGround("pl-1", well);
        TargetResolution.Result result = TargetResolution.resolve(ground, Map.of("type", "place", "place", "pl-1"));
        assertEquals(well, result.place());
    }

    @Test
    void noneTargetResolvesToNoPlace() {
        TargetResolution.Ground ground = stubGround("irrelevant", null);
        assertNull(TargetResolution.resolve(ground, Map.of("type", "none")).place());
        assertNull(TargetResolution.resolve(ground, null).place());
    }

    @Test
    void issuedTokenWithNoRecordedPlaceResolvesToNullPlace() {
        TargetResolution.Ground ground = stubGround("b-tam", null);
        TargetResolution.Result result = TargetResolution.resolve(ground, Map.of("type", "body", "body", "b-tam"));
        assertEquals(TargetResolution.Outcome.RESOLVED, result.outcome());
        assertNull(result.place());
    }
}
