package dev.kithcraft.mod.live;

import dev.kithcraft.mod.KithcraftMod;
import dev.kithcraft.mod.act.ClaimRegistry;
import dev.kithcraft.mod.act.IntentHandler;
import dev.kithcraft.mod.board.Board;
import dev.kithcraft.mod.board.BoardClaims;
import dev.kithcraft.mod.percept.PerceptEmitter;
import dev.kithcraft.mod.percept.SelfState;
import dev.kithcraft.mod.session.Continuity;
import dev.kithcraft.mod.session.Handshake;
import dev.kithcraft.mod.tokens.TokenRegistryData;
import dev.kithcraft.mod.wire.WireClient;
import net.minecraft.server.MinecraftServer;
import net.minecraft.world.entity.npc.villager.Villager;
import org.slf4j.Logger;

import java.io.IOException;
import java.nio.file.Path;
import java.util.Map;
import java.util.UUID;

/**
 * T011's minimal live wiring: one villager, one {@link WireClient} session dialed to the
 * mind daemon's UDS socket, real percepts out (a slow {@code self_state} heartbeat — proof
 * of card AC #1's live half) and real intents in (decoded and driven through {@link
 * IntentHandler} against a {@link LiveActuator}, proving AC #6 live). Single-body scope
 * deliberately: T011 is a dev-server *proof*, not the multi-body session manager V3 needs.
 *
 * <p>TASK-0020 T004: also wires the body's {@link ClaimRegistry} to the real {@link Board}
 * via {@link BoardClaims}, so a claim intent this body sends registers for real.
 *
 * <p>ponytail: the socket path is a bare system property ({@code kithcraft.socket}, default
 * {@code kithcraft.sock} relative to the server's run directory) — no config file, no
 * per-body multiplexing beyond what one connection already gives for free. Ceiling: fine
 * for a one-body dev-server proof; a real config surface and multi-body attach loop are
 * later work once more than one body vendor session needs managing at once.
 *
 * <p>TASK-0021 T003: {@link #villagerId()}/{@link #body()} expose the (UUID, body token) pair
 * this session actually attached to, so {@code KithcraftMod} can fold it into the same {@code
 * Function<UUID, Optional<String>>} lookup {@code DuskPairing#bodyTokenFor} already provides —
 * see {@link BodyTokenLookups} for why: {@code LiveBuildExecution#findClaimant} was consulting
 * only {@code DuskPairing}'s seat tokens, a disjoint namespace from the token a live claim
 * actually carries (this class's own {@code ground.issueBody} call above), so a live claim's
 * body could never be found (specs/020-job-board/research/board-observation.md §4).
 */
public final class BodySession {
    private static final Logger LOGGER = KithcraftMod.LOGGER;
    private static final long HEARTBEAT_INTERVAL_TICKS = 100; // ~5s at 20 TPS

    private final UUID villagerId;
    private final String body;
    private final PerceptEmitter emitter;
    private final IntentHandler intents;
    private long lastHeartbeat = Long.MIN_VALUE;

    private BodySession(UUID villagerId, String body, PerceptEmitter emitter, IntentHandler intents) {
        this.villagerId = villagerId;
        this.body = body;
        this.emitter = emitter;
        this.intents = intents;
    }

    /** The villager entity this session is attached to — the UUID half of the (UUID, body
     * token) pair a {@code bodyTokenLookup} resolves (T003). */
    public UUID villagerId() {
        return villagerId;
    }

    /** This session's own body token — the SAME token a live claim intent from this villager
     * actually carries (T003; see class javadoc). */
    public String body() {
        return body;
    }

    /** Dials the mind daemon, sends {@code session_open} for {@code villager}, and starts a
     * background reader that drives incoming {@code intent} messages through {@link
     * IntentHandler}. Throws if the dial fails (e.g. no mind daemon listening yet) — the
     * caller retries on a later tick. {@code board}/{@code boardThingToken} thread T004's
     * claim wiring through to this body's {@link IntentHandler} (see class javadoc). */
    public static BodySession open(MinecraftServer server, Villager villager, Board board,
            String boardThingToken) throws IOException {
        String socketPath = System.getProperty("kithcraft.socket", "kithcraft.sock");
        WireClient client = new WireClient(Path.of(socketPath));
        client.dial();

        TokenRegistryData tokens = server.overworld().getDataStorage().computeIfAbsent(TokenRegistryData.TYPE);
        LiveGround ground = new LiveGround(tokens);
        String bodyToken = ground.issueBody("a villager", villager.blockPosition());
        String placeToken = ground.issuePlace("a nearby spot", villager.blockPosition().offset(4, 0, 0));

        String sessionId = "s-live-" + UUID.randomUUID();
        long worldTime = server.overworld().getGameTime();
        Map<String, Object> sessionOpen =
            Handshake.buildSessionOpen(sessionId, bodyToken, worldTime, Continuity.firstSession());
        client.send(sessionOpen);
        LOGGER.info("[live] session_open sent: session={} body={} place={}", sessionId, bodyToken, placeToken);

        PerceptEmitter emitter = PerceptEmitter.overWireClient(client, sessionId);
        LiveActuator actuator = new LiveActuator(villager, ground);
        ClaimRegistry claims = new BoardClaims(board, boardThingToken, tokens::resolve);
        IntentHandler handler = new IntentHandler(emitter, ground, actuator, claims, sessionId, bodyToken);

        Thread reader = new Thread(() -> readLoop(client, handler), "kithcraft-live-reader");
        reader.setDaemon(true);
        reader.start();

        return new BodySession(villager.getUUID(), bodyToken, emitter, handler);
    }

    @SuppressWarnings("unchecked")
    private static void readLoop(WireClient client, IntentHandler handler) {
        try {
            Object msg;
            while ((msg = client.receive()) != null) {
                if (!(msg instanceof Map)) {
                    continue;
                }
                Map<String, Object> envelope = (Map<String, Object>) msg;
                if (!"intent".equals(envelope.get("message"))) {
                    continue;
                }
                Map<String, Object> payload = (Map<String, Object>) envelope.get("payload");
                long worldTime = envelope.get("world_time") instanceof Number n ? n.longValue() : 0L;
                Map<String, Object> ack = handler.handle(payload, worldTime);
                client.send(ack);
            }
        } catch (IOException e) {
            LOGGER.info("[live] reader ended: {}", e.toString());
        }
    }

    /** Called once per server tick: advances pending {@code go_to} progress and emits the
     * {@code self_state} heartbeat. */
    public void tick(long worldTime) throws IOException {
        intents.tick(worldTime);
        if (worldTime - lastHeartbeat >= HEARTBEAT_INTERVAL_TICKS) {
            lastHeartbeat = worldTime;
            emitter.emit(body, "self_state", PerceptEmitter.Urgency.BACKGROUND,
                SelfState.provenance(worldTime, worldTime), null,
                SelfState.content(SelfState.HUNGER, SelfState.Level.NONE, SelfState.Trend.STEADY),
                worldTime);
        }
    }
}
