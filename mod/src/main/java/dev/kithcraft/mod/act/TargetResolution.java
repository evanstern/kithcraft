package dev.kithcraft.mod.act;

import dev.kithcraft.mod.percept.Place;

import java.util.Map;
import java.util.Objects;
import java.util.Optional;

/**
 * §5.6: resolves an intent's {@code target} to ground truth without ever live-checking
 * existence. A {@code body}/{@code thing}/{@code person} target resolves to its LAST-SEEN
 * place (never a live position, PM-8); a {@code place} target is already ground truth.
 * {@code unknown_target} is refused ONLY when the named token was never issued (§5.3) — a
 * known-but-gone referent is ACCEPTED here and discovered only by {@link Verbs} trying the
 * act (§5.6's one sanctioned exception, {@code target_gone}).
 */
public final class TargetResolution {
    private TargetResolution() {}

    /** The only ground truth this resolver may consult: token issuance and the last-seen
     * place bookkeeping percept emission already gathers. Never a live position or a live
     * existence check — that is exactly the query §5.6 forbids. Production wiring (T011)
     * backs this with {@link dev.kithcraft.mod.tokens.TokenRegistry#resolve} plus a
     * last-seen-place tracker; tests substitute a scripted stub, the same pattern as {@link
     * dev.kithcraft.mod.percept.Sightings.HostileSource}. */
    public interface Ground {
        /** Whether {@code token} was ever issued (by any token type) — {@code false} here is
         * the ONLY legitimate {@code unknown_target} case (§5.3). */
        boolean isIssued(String token);

        /** The ground-truth place for an issued token: for a {@code place} token, that place
         * itself; for a {@code body}/{@code thing}/{@code person} token, the LAST-SEEN place
         * recorded for it (never live position, PM-8) — empty if the token is issued but no
         * place has ever been recorded for it. */
        Optional<Place> placeFor(String token);
    }

    public enum Outcome { RESOLVED, UNKNOWN_TARGET }

    /** @param place the resolved place; {@code null} for a {@code none} target, or a
     *     resolved body/thing/person target with no place ever recorded for it. */
    public record Result(Outcome outcome, Place place) {
        public static Result unknownTarget() {
            return new Result(Outcome.UNKNOWN_TARGET, null);
        }

        public static Result resolved(Place place) {
            return new Result(Outcome.RESOLVED, place);
        }
    }

    /** Resolves one {@code target} object per §5.2/§5.5: {@code {type: "place"|"thing"|
     * "body"|"person"|"none", ...}}. A {@code person} target resolves the same way as a
     * {@code thing} target — the protocol's own examples emit a person as a {@code k:person}
     * thing carrying a {@code thing_id} (§4.2, line ~322), not a distinct token namespace
     * (§2.3 names only {@code body}/{@code place}/{@code thing_id}/{@code kind}). A {@code
     * null} target (the {@code none} shorthand) resolves to no place. */
    public static Result resolve(Ground ground, Map<String, Object> target) {
        Objects.requireNonNull(ground, "ground");
        if (target == null) {
            return Result.resolved(null);
        }
        String type = (String) target.get("type");
        Objects.requireNonNull(type, "target.type");
        String token = switch (type) {
            case "none" -> null;
            case "place" -> (String) target.get("place");
            case "body" -> (String) target.get("body");
            case "thing", "person" -> (String) target.get("thing_id");
            default -> throw new IllegalArgumentException("unrecognized target type: " + type);
        };
        if (token == null) {
            return Result.resolved(null);
        }
        if (!ground.isIssued(token)) {
            return Result.unknownTarget();
        }
        return Result.resolved(ground.placeFor(token).orElse(null));
    }
}
