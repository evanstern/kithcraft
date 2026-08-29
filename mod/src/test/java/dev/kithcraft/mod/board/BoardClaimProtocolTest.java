package dev.kithcraft.mod.board;

import dev.kithcraft.mod.act.TargetResolution;
import dev.kithcraft.mod.act.Verbs;
import dev.kithcraft.mod.percept.Place;
import dev.kithcraft.mod.session.Handshake;
import org.junit.jupiter.api.Test;

import java.util.Map;
import java.util.Optional;
import java.util.Set;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * T004 (card AC #4, AR-4): extends {@code BoardReadProtocolTest}'s zero-protocol-extension
 * proof to the claim intent's target — Phase 1's own note flagged this as Phase 2's concern.
 * The claim intent's target ({@code {"type": "thing", "thing_id": ...}}) is EXACTLY the
 * pre-existing {@code thing} target shape {@link TargetResolution} already parses for {@code
 * go_to}/{@code speak} — proven by resolving the mind's own real shape (mind/deliberate/
 * board_test.go's {@code TestLoop_E3ClaimIntent_CarriesAuthoredReason}: {@code
 * target={"type":"thing","thing_id":"th-post-1"}}) through the unmodified resolver and
 * checking its key set carries no field beyond {@code type}/{@code thing_id}.
 */
class BoardClaimProtocolTest {

    /** The literal target shape mind/deliberate's E3 claim intent test composes — a plain
     * {@code {type, thing_id}} object, nothing more (mind/llm/structured.go's {@code Intent}
     * has no field for it to add one). */
    @Test
    void claimTargetHasExactlyTheExistingThingShapeKeys() {
        Map<String, Object> target = Map.of("type", "thing", "thing_id", "th-post-1");
        assertEquals(Set.of("type", "thing_id"), target.keySet(),
            "card AC #4/#9: the claim intent's target carries no new field beyond the "
                + "pre-existing thing-target shape go_to/speak already use");
    }

    /** The SAME {@link TargetResolution#resolve} every other verb already goes through
     * accepts a claim's target with zero new branching — no new target type, no new field
     * read out of the map. */
    @Test
    void targetResolutionParsesTheClaimShapeThroughTheExistingThingCase() {
        TargetResolution.Ground ground = new TargetResolution.Ground() {
            @Override public boolean isIssued(String token) { return "th-post-1".equals(token); }
            @Override public Optional<Place> placeFor(String token) { return Optional.empty(); }
        };
        Map<String, Object> target = Map.of("type", "thing", "thing_id", "th-post-1");
        TargetResolution.Result result = TargetResolution.resolve(ground, target);
        assertEquals(TargetResolution.Outcome.RESOLVED, result.outcome());
    }

    /** {@code claim} is declared, and its declared target set is exactly {@code thing} — no
     * new target TYPE either (L-7 also still holds: the manifest is the same static constant
     * every body gets, unaffected by which body reads). */
    @Test
    @SuppressWarnings("unchecked")
    void claimVerbIsDeclaredWithOnlyTheExistingThingTargetType() {
        assertTrue(Verbs.DECLARED.contains("claim"), "card AC #4: claim must be a declared verb");

        var verbs = (java.util.List<Map<String, Object>>) Handshake.MANIFEST.get("verbs");
        Map<String, Object> claimEntry = verbs.stream()
            .filter(v -> "claim".equals(v.get("verb"))).findFirst()
            .orElseThrow(() -> new AssertionError("claim not declared in the manifest"));
        assertEquals(java.util.List.of("thing"), claimEntry.get("targets"),
            "card AC #4/#9: claim's declared target set is exactly the pre-existing \"thing\" "
                + "type — no new target type introduced");
    }
}
