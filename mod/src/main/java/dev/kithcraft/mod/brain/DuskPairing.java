package dev.kithcraft.mod.brain;

import dev.kithcraft.mod.KithcraftMod;
import dev.kithcraft.mod.cast.Cast;
import dev.kithcraft.mod.tokens.TokenRegistry;
import dev.kithcraft.mod.tokens.TokenRegistryData;
import net.minecraft.core.BlockPos;
import net.minecraft.core.GlobalPos;
import net.minecraft.server.level.ServerLevel;
import net.minecraft.world.entity.ai.memory.MemoryModuleType;
import net.minecraft.world.entity.ai.village.poi.PoiTypes;
import net.minecraft.world.entity.npc.villager.Villager;
import net.minecraft.world.entity.schedule.Activity;
import net.minecraft.world.level.block.Blocks;
import net.minecraft.world.level.pathfinder.Path;
import net.minecraft.world.phys.Vec3;
import org.slf4j.Logger;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.OptionalDouble;
import java.util.UUID;

/**
 * T007 (FR-005, R-7): dusk pair formation, driven by vanilla's own MEET machinery rather
 * than a new Activity or Mixin — the Mixin budget stays at 2 (research/brain-26.2.md §4/§5's
 * own flagged option, and the dispatch's stated preference). {@code
 * VillagerGoalPackages.getMeetPackage(float)} already contains {@code
 * StrollAroundPoi(MemoryModuleType.MEETING_POINT, ...)} and {@code SocializeAtBell.create()}
 * (confirmed by {@code javap -p -c} against the exact byte range of that one method,
 * 2026-08-27: no {@code VillagerMakeLove} reference in it — breeding stays confined to
 * {@code getIdlePackage}, already suppressed live by {@code VillagerGoalPackagesMixin}).
 * Both {@code PoiTypes.MEETING} and {@code MemoryModuleType.MEETING_POINT} are public
 * (research §3), so placing one Bell (a vanilla "meeting" POI) and setting {@code
 * MEETING_POINT} directly on each cast member's Brain gets the whole convergence for
 * free — no new Activity, no new Mixin, no custom BehaviorControl. Setting the memory
 * directly (rather than waiting on vanilla's own POI-sensing to claim it) mirrors {@link
 * ScheduleSetup}'s {@code JOB_SITE} precedent, for the identical PoiManager-registration-race
 * reason documented there.
 *
 * <p>This class does not path anyone itself — vanilla's MEET behaviors do. It only (a)
 * builds the shared gathering place once and (b) watches each member's own {@code
 * PathNavigation} every tick to fire {@link PairingSignal} ~10s before the slower of a
 * candidate pair arrives.
 *
 * <p>ponytail: "pick a pair" takes the first two cast members currently active in MEET with
 * a live path, in {@link Cast#MEMBERS} order — with exactly three members a third
 * simultaneous arrival isn't excluded, and the signal only ever describes the chosen two.
 * Ceiling: fine for a 3-member cast; a real pairing policy (who hasn't talked recently, etc.)
 * is M6/persona territory, not this vendor-side signal.
 *
 * <p>ponytail: the gathering place's token is reissued every server restart (this class is
 * built fresh in {@code KithcraftMod.onServerStarted} each boot) rather than persisted —
 * unlike cast identity (FR-002), the spec places no restart-survival requirement on the
 * gathering place, so no new persisted-string plumbing was added for it (plan.md: "no new
 * persistence design"). Harmless: {@link TokenRegistry} tokens are never reused, so an extra
 * token per restart is inert, not a bug.
 *
 * <p>ponytail: firing logs the composed {@code sighting} content (T009's dev-server
 * evidence) but does not yet send it over a live {@code WireClient} session — {@code
 * BodySession}'s own doc already flags this mod's live wiring as single-body only. Composing
 * the correct percept through the correct existing type is this task's job; multiplexing it
 * across two live mind sessions is the session-manager work that ponytail defers.
 *
 * <p><b>Live finding (T009):</b> {@code StrollAroundPoi}/{@code SocializeAtBell} (the vanilla
 * behaviors {@code getMeetPackage} drives {@code MEETING_POINT} through) loiter — repeated
 * short hops near the bell, not one continuous walk from far away — so {@code
 * PathNavigation.getPath()} goes briefly {@code null}/done between every hop. An earlier
 * version reset {@link PairingSignal}'s dedup on every such gap, which re-armed the signal
 * once per hop and fired it dozens of times at small etas instead of once ~10s before the
 * real first arrival. Fixed two ways: (1) a momentary path gap no longer resets dedup, only
 * skips that tick (dedup survives the gap between hops); (2) a pair already within {@link
 * #ARRIVAL_RADIUS_BLOCKS} of the gathering place is treated as already met — no more firing,
 * and dedup only resets there, so a later, genuinely fresh approach (next dusk) can fire
 * again. Confirmed live: a single clean fire per approach, ETA within the lead window,
 * followed by real convergence — not a firehose of small-eta repeats.
 */
public final class DuskPairing {
    private static final Logger LOGGER = KithcraftMod.LOGGER;

    /** A pair within this many blocks of the gathering place is treated as already met —
     * vanilla's own stroll/socialize radius is a few blocks, so this is comfortably inside
     * it (T009 live finding, see class javadoc). */
    private static final double ARRIVAL_RADIUS_BLOCKS = 4.0;

    private record Seat(Villager villager, Cast.Member member, String bodyToken) {
    }

    private final String placeToken;
    private final BlockPos placePos;
    private final List<Seat> seats;
    private final PairingSignal signal = new PairingSignal();

    private DuskPairing(String placeToken, BlockPos placePos, List<Seat> seats) {
        this.placeToken = placeToken;
        this.placePos = placePos;
        this.seats = seats;
    }

    /** Builds the shared gathering place (one Bell, registered as the vanilla {@code
     * meeting} POI) near {@code origin}, issues its place token and each matched cast
     * member's body token via the existing {@link TokenRegistry}, and sets {@code
     * MEETING_POINT} directly on every matched member's Brain. Call once per server start;
     * re-placing the same bell / re-setting the same memory is a no-op in effect. Any {@code
     * castVillagers} entry whose custom name doesn't match {@link Cast#MEMBERS} is skipped. */
    public static DuskPairing setUp(ServerLevel level, TokenRegistryData tokens, BlockPos origin,
            List<Villager> castVillagers) {
        BlockPos bellPos = origin.offset(Cast.MEMBERS.size() + 2, 0, 0);
        var bellState = Blocks.BELL.defaultBlockState();
        level.setBlockAndUpdate(bellPos, bellState);
        if (level.getPoiManager().getType(bellPos).isEmpty()) {
            PoiTypes.forState(bellState).ifPresent(poiType -> level.getPoiManager().add(bellPos, poiType));
        }
        GlobalPos gatheringPlace = GlobalPos.of(level.dimension(), bellPos);
        String placeToken = tokens.issue(TokenRegistry.TokenType.PLACE, "the village gathering place");

        Map<String, Cast.Member> byName = new LinkedHashMap<>();
        Cast.MEMBERS.forEach(m -> byName.put(m.name(), m));

        List<Seat> seats = new ArrayList<>();
        for (Villager villager : castVillagers) {
            Cast.Member member = byName.get(villager.getName().getString());
            if (member == null) {
                continue;
            }
            villager.getBrain().setMemory(MemoryModuleType.MEETING_POINT, gatheringPlace);
            String bodyToken = tokens.issue(TokenRegistry.TokenType.BODY, member.name());
            LOGGER.info("[dusk] {} body token: {}", member.name(), bodyToken);
            seats.add(new Seat(villager, member, bodyToken));
        }
        return new DuskPairing(placeToken, bellPos, List.copyOf(seats));
    }

    /** Called once per server tick. Finds every candidate pair among members currently
     * active in {@code Activity.MEET}, estimates each pair's arrival from the members' own
     * navigation paths, and fires {@link PairingSignal} ~10s ahead (R-7). */
    public void tick(long worldTime) {
        List<Seat> approaching = seats.stream()
            .filter(s -> s.villager().getBrain().isActive(Activity.MEET))
            .toList();
        for (int i = 0; i < approaching.size(); i++) {
            for (int j = i + 1; j < approaching.size(); j++) {
                considerPair(approaching.get(i), approaching.get(j), worldTime);
            }
        }
    }

    private void considerPair(Seat a, Seat b, long worldTime) {
        String pairKey = a.bodyToken() + "+" + b.bodyToken();
        if (withinArrivalRadius(a.villager()) && withinArrivalRadius(b.villager())) {
            // Already met (T009 live finding: vanilla's own loitering keeps both nearby for
            // the rest of MEET) — stop firing, and clear dedup so a later, genuinely fresh
            // approach (e.g. next dusk) can fire again.
            signal.reset(pairKey);
            return;
        }
        OptionalDouble etaA = eta(a.villager());
        OptionalDouble etaB = eta(b.villager());
        OptionalDouble pairEta = PairingSignal.combine(etaA, etaB);
        if (pairEta.isEmpty()) {
            // R-7's no-fire edge case, OR just the momentary gap between two of vanilla's
            // own short stroll hops (T009 live finding) — skip this tick only, without
            // clearing dedup, so a real pathing failure doesn't refire the same approach a
            // moment later once a fresh hop's path happens to compute.
            return;
        }
        if (signal.shouldFireNow(pairKey, pairEta)) {
            LOGGER.info("[dusk] pairing signal: {} + {} at {} (eta {}s, world_time {})",
                a.member().name(), b.member().name(), placeToken, pairEta.getAsDouble(), worldTime);
            LOGGER.info("[dusk] {} perceives: {}", a.member().name(),
                PairingSignal.content(b.bodyToken(), b.member().name()));
            LOGGER.info("[dusk] {} perceives: {}", b.member().name(),
                PairingSignal.content(a.bodyToken(), a.member().name()));
        }
    }

    private boolean withinArrivalRadius(Villager villager) {
        return villager.blockPosition().distSqr(placePos) <= ARRIVAL_RADIUS_BLOCKS * ARRIVAL_RADIUS_BLOCKS;
    }

    /** Remaining path distance (sum of the still-unwalked node segments, plus the entity's
     * own distance to the next node) over the entity's current AI movement speed — the
     * "path-length/speed estimate from the villagers' actual navigation paths" T008 asks
     * for. Empty when there is no active path, the path is done, or navigation is stuck
     * (pathing failure, the spec's no-fire edge case). */
    private static OptionalDouble eta(Villager villager) {
        Path path = villager.getNavigation().getPath();
        if (path == null || path.isDone() || villager.getNavigation().isStuck()) {
            return OptionalDouble.empty();
        }
        int next = path.getNextNodeIndex();
        int count = path.getNodeCount();
        if (next >= count) {
            return OptionalDouble.empty();
        }
        Vec3 entityPos = villager.position();
        double remaining = entityPos.distanceTo(path.getEntityPosAtNode(villager, next));
        for (int i = next; i < count - 1; i++) {
            remaining += path.getNode(i).distanceTo(path.getNode(i + 1));
        }
        double blocksPerSecond = villager.getSpeed() * 20.0;
        return PairingSignal.predictedArrivalSeconds(OptionalDouble.of(remaining), blocksPerSecond);
    }

    public String placeToken() {
        return placeToken;
    }

    /** T007 (card AC #9): the cast member's own body token, by entity UUID — the lookup
     * {@code dev.kithcraft.mod.live.LiveDeathHandling} needs to retire the right token on
     * death. Empty for any villager this pairing never matched/seated (not a cast member, or
     * seated before this pairing was set up). */
    public Optional<String> bodyTokenFor(UUID villagerId) {
        return seats.stream()
            .filter(s -> s.villager().getUUID().equals(villagerId))
            .map(Seat::bodyToken)
            .findFirst();
    }
}
