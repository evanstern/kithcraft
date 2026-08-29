package dev.kithcraft.mod.board;

import dev.kithcraft.mod.KithcraftMod;
import dev.kithcraft.mod.percept.PerceptEmitter;
import dev.kithcraft.mod.percept.Place;
import dev.kithcraft.mod.percept.Provenance;
import dev.kithcraft.mod.percept.Testimony;
import net.minecraft.core.BlockPos;
import net.minecraft.world.entity.npc.villager.Villager;
import net.minecraft.world.entity.schedule.Activity;
import org.slf4j.Logger;

import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.function.Function;

/**
 * T002 (FR-002, card ACs #2, #3): the board's read cycle. Rides the exact substrate {@code
 * dev.kithcraft.mod.brain.DuskPairing} already established for dusk convergence —
 * {@code Activity.MEET} (villager-brain-api's addActivity-additive facts: no new Activity, no
 * new Mixin) — rather than minting any board-specific AI: {@link BoardSetup} places the board
 * inside the same gathering radius MEET already walks cast members into, so this class only
 * watches proximity each tick and fires the read once per approach (mirrors {@code
 * DuskPairing}'s own arrival-radius + dedup shape, via {@link BoardReadTracker}).
 *
 * <p>The percept itself is composed through the SAME §4.7 machinery V2 already built —
 * {@link Testimony#textContent}/{@link Testimony#textProvenance}, {@link PerceptEmitter} —
 * with no new field, shape, or percept type: this is card AC #3's "zero protocol extension"
 * proof, checked structurally in {@code BoardReadProtocolTest}.
 *
 * <p>ponytail: like {@code DuskPairing}, this composes and logs the read (dev-server evidence)
 * rather than pushing it over a live multi-body {@code WireClient} session — {@code
 * BodySession} is single-body scoped (see its own doc); multiplexing a shared percept across
 * several live mind sessions is the session-manager work {@code DuskPairing}'s own ponytail
 * already deferred, not new scope this task invents. Ceiling: fine for the Phase 1/4
 * dev-server proof; wire through a real {@link PerceptEmitter} sink once multi-body sessions
 * exist.
 */
public final class BoardVisit {
    private static final Logger LOGGER = KithcraftMod.LOGGER;
    private static final double ARRIVAL_RADIUS_BLOCKS = 4.0;

    private final BlockPos boardPos;
    private final String boardThingToken;
    private final Place boardPlace;
    private final Board board;
    private final BoardReadTracker tracker = new BoardReadTracker();

    public BoardVisit(BlockPos boardPos, String boardThingToken, Place boardPlace, Board board) {
        this.boardPos = boardPos;
        this.boardThingToken = boardThingToken;
        this.boardPlace = boardPlace;
        this.board = board;
    }

    /** Called once per server tick. {@code bodyTokenLookup} resolves a villager's own body
     * token (reuses whatever already issued it — {@code DuskPairing#bodyTokenFor}; this class
     * mints no new tokens); a villager with no resolvable token is skipped (never seated, or
     * seated by a later pairing setup than this tick's). */
    public void tick(long worldTime, List<Villager> villagers, Function<UUID, Optional<String>> bodyTokenLookup) {
        for (Villager villager : villagers) {
            boolean near = villager.getBrain().isActive(Activity.MEET)
                && villager.blockPosition().distSqr(boardPos) <= ARRIVAL_RADIUS_BLOCKS * ARRIVAL_RADIUS_BLOCKS;
            if (tracker.shouldFireNow(villager.getStringUUID(), near)) {
                bodyTokenLookup.apply(villager.getUUID())
                    .ifPresent(bodyToken -> fireRead(bodyToken, villager.getName().getString(), worldTime));
            }
        }
    }

    private void fireRead(String bodyToken, String villagerName, long worldTime) {
        Provenance.Source source = Provenance.Source.artifact(boardThingToken, "the job board");
        Provenance provenance = Testimony.textProvenance(source, null, worldTime);
        Map<String, Object> content = board.readableContent();

        PerceptEmitter emitter = new PerceptEmitter(m -> { }, "s-board-log");
        Map<String, Object> envelope = emitter.compose(bodyToken, "text", PerceptEmitter.Urgency.NOTABLE,
            provenance, boardPlace, content, worldTime);

        LOGGER.info("[board] {} reads the board: {}", villagerName, envelope);
    }
}
