package dev.kithcraft.mod.tokens;

import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;

/**
 * Issues, resolves, and retires the four token types the protocol defines
 * (docs/design/body-protocol-v0.md §2.3): {@code body}, {@code place}, {@code thing_id},
 * {@code kind}. Pure core — no Minecraft/Fabric types — so it is unit-testable without a
 * live server; {@link TokenRegistryData} is the thin adapter that persists a {@link
 * Snapshot} through the world save (T008).
 *
 * <p>The referent is an opaque descriptor string; binding a token to an actual world object
 * (an entity, a position, a POI) is V2's job — this skeleton only guarantees the identity
 * bookkeeping. §2.3's hardest obligation — a token, once issued, is <b>never</b> reissued for
 * a different referent — is enforced structurally: a monotonic counter only ever grows, is
 * carried whole by {@link #snapshot()}/{@link #restore(Snapshot)}, and never resets, so no
 * token string this registry has ever produced can be produced again.
 */
public final class TokenRegistry {

    /** The four token namespaces (§2.3), each spelled with the protocol's own prefix
     * (e.g. {@code b-1}, {@code pl-2}, {@code th-3}, {@code k-4}). */
    public enum TokenType {
        BODY("b"), PLACE("pl"), THING("th"), KIND("k");

        private final String prefix;

        TokenType(String prefix) {
            this.prefix = prefix;
        }
    }

    /** One issued token's persisted state. Public so {@link TokenRegistryData} can build a
     * codec over it without this class exposing any richer internal representation. */
    public record Entry(String referent, boolean retired) {
        public Entry {
            Objects.requireNonNull(referent, "referent");
        }
    }

    /** A pure-JDK snapshot of the registry's full state — the monotonic counter and every
     * entry ever issued, including retired ones, so retirement itself survives a restart. */
    public record Snapshot(long nextSequence, Map<String, Entry> entries) {
        public Snapshot {
            entries = Map.copyOf(entries);
        }
    }

    private long nextSequence;
    private final Map<String, Entry> entries = new LinkedHashMap<>();

    public TokenRegistry() {
        this.nextSequence = 0L;
    }

    private TokenRegistry(long nextSequence, Map<String, Entry> entries) {
        this.nextSequence = nextSequence;
        this.entries.putAll(entries);
    }

    /** Issues a new, never-before-used token of the given type for {@code referent}. */
    public synchronized String issue(TokenType type, String referent) {
        Objects.requireNonNull(type, "type");
        Objects.requireNonNull(referent, "referent");
        String token = type.prefix + "-" + nextSequence++;
        entries.put(token, new Entry(referent, false));
        return token;
    }

    /** The token's referent, or empty if the token was never issued or has since been
     * retired. */
    public synchronized Optional<String> resolve(String token) {
        Entry e = entries.get(token);
        return (e == null || e.retired()) ? Optional.empty() : Optional.of(e.referent());
    }

    /** Retires a token: it stops resolving. Because issuance never rewinds the counter, the
     * token string itself is never produced again regardless (§2.3's never-reuse rule). */
    public synchronized void retire(String token) {
        Entry e = entries.get(token);
        if (e != null) {
            entries.put(token, new Entry(e.referent(), true));
        }
    }

    public synchronized Snapshot snapshot() {
        return new Snapshot(nextSequence, entries);
    }

    /** All tokens that currently resolve (not retired), token → referent. For diagnostics
     * and the dev-server restart observation (T009) — not part of any protocol surface. */
    public synchronized Map<String, String> liveEntries() {
        Map<String, String> out = new LinkedHashMap<>();
        entries.forEach((token, e) -> {
            if (!e.retired()) {
                out.put(token, e.referent());
            }
        });
        return out;
    }

    public static TokenRegistry restore(Snapshot snapshot) {
        Objects.requireNonNull(snapshot, "snapshot");
        return new TokenRegistry(snapshot.nextSequence(), snapshot.entries());
    }
}
