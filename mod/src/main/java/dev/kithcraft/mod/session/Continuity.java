package dev.kithcraft.mod.session;

/**
 * {@code session_open}'s continuity fields (protocol §6.2, seam-wire-v0.md §1.5). Reports
 * <em>that</em> a body was unattended and for how long — never what happened during the
 * gap (the vendor MUST NOT backfill it as though the body had been watching).
 *
 * <p>Matching a returning body to its previous session is the token registry's job
 * (Phase 3, T008: a mind matches by the stable {@code body} token, never by
 * {@code previous_session}). Until that registry exists, every body opens as
 * {@link #firstSession()}.
 */
public record Continuity(String previousSession, Long previousCloseWorldTime, boolean bodyContinuous) {

    /** No prior session to report — the case until the token registry (Phase 3) can look
     * one up for a given body. */
    public static Continuity firstSession() {
        return new Continuity(null, null, false);
    }
}
