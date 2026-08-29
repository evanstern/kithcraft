package dev.kithcraft.mod.board;

import dev.kithcraft.mod.act.ClaimRegistry;

import java.util.Objects;
import java.util.Optional;
import java.util.function.Function;

/**
 * T004: the engine-side {@link ClaimRegistry} wiring — resolves a claim intent's target
 * token against this board's own thing token (the one every read percept already carries in
 * its {@code provenance.source}, {@link BoardVisit}) and, on a match, registers the claim on
 * {@link Board} with the claimant's own descriptor rather than its bare body token, so the
 * board's claim line reads like a person claimed it. {@code descriptorOf} is a lookup into
 * whatever issued the claimant's body token — real wiring is {@code
 * TokenRegistryData::resolve} (the villager's own name, the same referent {@code DuskPairing}
 * issued the token under); tests substitute a plain map lookup.
 */
public final class BoardClaims implements ClaimRegistry {
    private final Board board;
    private final String boardThingToken;
    private final Function<String, Optional<String>> descriptorOf;

    public BoardClaims(Board board, String boardThingToken, Function<String, Optional<String>> descriptorOf) {
        this.board = Objects.requireNonNull(board, "board");
        this.boardThingToken = Objects.requireNonNull(boardThingToken, "boardThingToken");
        this.descriptorOf = Objects.requireNonNull(descriptorOf, "descriptorOf");
    }

    @Override
    public Outcome claim(String claimantBody, String thingToken) {
        if (!boardThingToken.equals(thingToken)) {
            return Outcome.UNKNOWN_POSTING;
        }
        String descriptor = descriptorOf.apply(claimantBody).orElse(claimantBody);
        return board.tryClaim(claimantBody, descriptor) ? Outcome.CLAIMED : Outcome.ALREADY_CLAIMED;
    }
}
