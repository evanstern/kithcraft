package dev.kithcraft.mod.live;

import dev.kithcraft.mod.act.TargetResolution;
import dev.kithcraft.mod.percept.Place;
import dev.kithcraft.mod.tokens.TokenRegistry;
import dev.kithcraft.mod.tokens.TokenRegistryData;
import net.minecraft.core.BlockPos;

import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;

/**
 * The live dev-server {@link TargetResolution.Ground} (T011): token issuance goes through
 * the real, save-persisted {@link TokenRegistryData}; a token's real {@link BlockPos} lives
 * only in this in-memory side table, since {@link Place} never carries coordinates on the
 * wire (AR-4) — {@link LiveActuator} is the only reader.
 *
 * <p>ponytail: a body token's position is a static snapshot taken when {@link #issueBody}
 * is called, not a continuously-updated last-seen place fed by real sighting percepts
 * (Sightings.java composes those; nothing yet drives them off a live nearby-entity scan).
 * Ceiling: correct and sufficient for T011's single-body dev-server proof, which targets
 * `go_to` at a place token; wire a live Sightings scan before trusting body-target
 * resolution past this phase.
 */
public final class LiveGround implements TargetResolution.Ground {
    private final TokenRegistryData tokens;
    private final Map<String, BlockPos> placePositions = new LinkedHashMap<>();
    private final Map<String, BlockPos> bodyLastSeen = new LinkedHashMap<>();

    public LiveGround(TokenRegistryData tokens) {
        this.tokens = Objects.requireNonNull(tokens, "tokens");
    }

    /** Issues a place token bound to a real position. */
    public String issuePlace(String referent, BlockPos pos) {
        String token = tokens.issue(TokenRegistry.TokenType.PLACE, referent);
        placePositions.put(token, pos);
        return token;
    }

    /** Issues a body token with its attach-time position as the initial last-seen place. */
    public String issueBody(String referent, BlockPos pos) {
        String token = tokens.issue(TokenRegistry.TokenType.BODY, referent);
        bodyLastSeen.put(token, pos);
        return token;
    }

    /** The real {@link BlockPos} behind a token — {@link LiveActuator}'s only way to turn a
     * resolved {@link Place} back into somewhere to path to. */
    public Optional<BlockPos> blockPosFor(String token) {
        BlockPos pos = placePositions.get(token);
        return Optional.ofNullable(pos != null ? pos : bodyLastSeen.get(token));
    }

    @Override
    public boolean isIssued(String token) {
        return tokens.resolve(token).isPresent();
    }

    @Override
    public Optional<Place> placeFor(String token) {
        if (!placePositions.containsKey(token) && !bodyLastSeen.containsKey(token)) {
            return Optional.empty();
        }
        String descriptor = tokens.resolve(token).orElse(null);
        return Optional.of(new Place(token, descriptor));
    }
}
