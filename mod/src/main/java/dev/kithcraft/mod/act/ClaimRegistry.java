package dev.kithcraft.mod.act;

/**
 * T004 (FR-... V4 US2, card AC #4; AR-4): resolves a {@code claim} intent's target token
 * against whatever posting it names, on behalf of the claiming body. This is AR-4's "the
 * mind names a token, the engine resolves it" applied to claiming exactly as {@link
 * TargetResolution.Ground} applies it to movement — {@link IntentHandler} never decides
 * who wins a claim, it only asks this seam and reports the result as an {@code act_result}
 * (§5.4).
 *
 * <p>Real wiring binds this to {@code dev.kithcraft.mod.board.Board} via {@code
 * dev.kithcraft.mod.board.BoardClaims} (constructed once per server start, threaded into
 * every body's {@link IntentHandler}); tests substitute a scripted stub — the same pattern
 * {@link TargetResolution.Ground} and {@link Verbs.Actuator} already established.
 */
public interface ClaimRegistry {

    enum Outcome {
        /** This body is now the claimant — first accepted claim wins (spec Edge Cases). */
        CLAIMED,
        /** Someone else already holds this posting; this body's claim loses. */
        ALREADY_CLAIMED,
        /** {@code thingToken} is an issued token this registry does not recognize as any
         * postable board — accepted at ack (it WAS issued, §5.3), failed here instead, the
         * same accepted-then-failed posture §5.6 uses for a gone referent. */
        UNKNOWN_POSTING
    }

    /** Attempts to claim the posting named by {@code thingToken} on behalf of {@code
     * claimantBody}. */
    Outcome claim(String claimantBody, String thingToken);
}
