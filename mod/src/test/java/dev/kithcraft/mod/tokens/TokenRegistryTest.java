package dev.kithcraft.mod.tokens;

import org.junit.jupiter.api.Test;

import java.util.Optional;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotEquals;

/**
 * T009: proves the pure {@link TokenRegistry} core — issue/resolve/retire, never-reuse, and
 * serialize → deserialize → same referents with the counter preserved (card AC #5's unit
 * half; the dev-server restart observation is the other half, recorded in
 * specs/009-fabric-mod-skeleton/research/).
 */
class TokenRegistryTest {

    @Test
    void issueThenResolveReturnsTheSameReferent() {
        TokenRegistry registry = new TokenRegistry();

        String body = registry.issue(TokenRegistry.TokenType.BODY, "villager Tam");
        String place = registry.issue(TokenRegistry.TokenType.PLACE, "the well");

        assertEquals(Optional.of("villager Tam"), registry.resolve(body));
        assertEquals(Optional.of("the well"), registry.resolve(place));
        assertNotEquals(body, place, "distinct tokens for distinct referents");
    }

    @Test
    void resolvingAnUnknownTokenIsEmpty() {
        TokenRegistry registry = new TokenRegistry();
        assertEquals(Optional.empty(), registry.resolve("b-999"));
    }

    @Test
    void retiredTokenStopsResolvingButIsNeverReissued() {
        TokenRegistry registry = new TokenRegistry();
        String token = registry.issue(TokenRegistry.TokenType.THING, "a bed");

        registry.retire(token);
        assertEquals(Optional.empty(), registry.resolve(token), "a retired token must not resolve");

        // The monotonic counter never rewinds, so no later issuance of the same type can
        // ever reproduce this exact token string for a different referent (§2.3).
        for (int i = 0; i < 5; i++) {
            String next = registry.issue(TokenRegistry.TokenType.THING, "referent " + i);
            assertNotEquals(token, next, "a retired token must never be reissued");
        }
    }

    @Test
    void snapshotRoundTripPreservesReferentsAndCounter() {
        TokenRegistry original = new TokenRegistry();
        String live = original.issue(TokenRegistry.TokenType.BODY, "villager Tam");
        String retired = original.issue(TokenRegistry.TokenType.PLACE, "the old well");
        original.retire(retired);

        TokenRegistry restored = TokenRegistry.restore(original.snapshot());

        assertEquals(Optional.of("villager Tam"), restored.resolve(live));
        assertEquals(Optional.empty(), restored.resolve(retired), "retirement itself survives the round trip");

        // The counter survived too: two tokens were issued before the snapshot (seq 0, 1),
        // so the next issuance after restore must continue at seq 2, not restart at 0.
        String afterRestore = restored.issue(TokenRegistry.TokenType.BODY, "villager Nell");
        assertEquals("b-2", afterRestore, "the monotonic counter must survive the round trip");
    }
}
