package dev.kithcraft.mod.live;

import dev.kithcraft.mod.KithcraftMod;
import dev.kithcraft.mod.board.BoardData;
import dev.kithcraft.mod.brain.DuskPairing;
import dev.kithcraft.mod.death.GraveBoardEntry;
import dev.kithcraft.mod.death.GravePlacement;
import dev.kithcraft.mod.death.GriefPeriod;
import dev.kithcraft.mod.percept.Place;
import dev.kithcraft.mod.percept.Thing;
import dev.kithcraft.mod.tokens.TokenRegistry;
import dev.kithcraft.mod.tokens.TokenRegistryData;
import net.fabricmc.fabric.api.entity.event.v1.ServerLivingEntityEvents;
import net.minecraft.core.BlockPos;
import net.minecraft.core.GlobalPos;
import net.minecraft.network.chat.Component;
import net.minecraft.server.level.ServerLevel;
import net.minecraft.world.SimpleContainer;
import net.minecraft.world.damagesource.DamageSource;
import net.minecraft.world.entity.LivingEntity;
import net.minecraft.world.entity.ai.memory.MemoryModuleType;
import net.minecraft.world.entity.ai.village.poi.PoiManager;
import net.minecraft.world.entity.npc.villager.Villager;
import net.minecraft.world.item.ItemStack;
import net.minecraft.world.level.block.Blocks;
import net.minecraft.world.level.block.entity.ChestBlockEntity;
import net.minecraft.world.level.block.entity.SignBlockEntity;
import org.slf4j.Logger;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.function.Supplier;

/**
 * T007-T009 (FR-004/005/006, card ACs #4/#5/#6/#7/#9): the live wiring for both death kinds
 * (ordinary kills and conversion-cancelled deaths, R-6 — both fall through {@code
 * Villager.die()} identically, so this class has no separate conversion branch).
 *
 * <p><b>No new Mixin.</b> Fabric API's own {@code ServerLivingEntityEvents} (already-shipped,
 * already-Mixin'd inside {@code fabric-entity-events-v1}, a dependency this mod already
 * declares transitively via {@code fabric-api}) is the capture hook: {@code ALLOW_DEATH} fires
 * at the very head of {@code LivingEntity.die()} — before {@code dropAllDeathLoot} or any other
 * loot/removal step runs — so the hidden inventory is copied out while it is still intact.
 * {@code AFTER_DEATH} fires once {@code die()} has fully run; by then vanilla has already
 * destroyed the hidden inventory (death-mechanics.md §3), but {@code Villager.releasePoi}
 * (R-4) never erases the {@code HOME}/{@code JOB_SITE} brain memory it reads — only the {@code
 * PoiManager} ticket — so those {@code GlobalPos} values are still readable here, which is what
 * makes the grief-period re-claim (T008) possible without a Mixin either.
 *
 * <p>ponytail: the grave sighting and belongings-thing percept content is composed and logged
 * here exactly the way {@code DuskPairing} logs {@code PairingSignal} content — not yet pushed
 * over a live {@code WireClient} session. Multiplexing percepts across more than one attached
 * body is {@code BodySession}'s already-flagged single-body ceiling, not new scope this class
 * should carry; T010 (Phase 4) is where percept-channel delivery gets its live proof. The board
 * posting itself is the exception (TASK-0020 T006): it rides the real, persisted {@code Board}
 * for real, not just a log line — the next board read genuinely carries it.
 */
public final class LiveDeathHandling {
    private static final Logger LOGGER = KithcraftMod.LOGGER;
    private static final Set<MemoryModuleType<GlobalPos>> GRIEF_POI_MEMORIES =
        Set.of(MemoryModuleType.HOME, MemoryModuleType.JOB_SITE);

    private final TokenRegistryData tokens;
    private final Supplier<DuskPairing> duskPairing;
    private final BoardData boardData;
    private final Map<UUID, List<ItemStack>> pendingCapture = new ConcurrentHashMap<>();
    private final List<PendingHold> holds = new ArrayList<>();

    private record PendingHold(BlockPos pos, GriefPeriod.Hold hold) {
    }

    private LiveDeathHandling(TokenRegistryData tokens, Supplier<DuskPairing> duskPairing, BoardData boardData) {
        this.tokens = tokens;
        this.duskPairing = duskPairing;
        this.boardData = boardData;
    }

    /** Registers the Fabric API death hooks once per server start. {@code boardData} is
     * TASK-0020 T006's seam closure: the grave posting this class composes on death now
     * rides {@code boardData}'s real {@code Board} instead of only being logged. */
    public static LiveDeathHandling register(
            TokenRegistryData tokens, Supplier<DuskPairing> duskPairing, BoardData boardData) {
        LiveDeathHandling handling = new LiveDeathHandling(tokens, duskPairing, boardData);
        ServerLivingEntityEvents.ALLOW_DEATH.register(handling::onAllowDeath);
        ServerLivingEntityEvents.AFTER_DEATH.register(handling::onAfterDeath);
        return handling;
    }

    /** Never vetoes death — the return value only exists because {@code ALLOW_DEATH} is a gate
     * event; this listener's only job is the pre-loot capture side effect. */
    private boolean onAllowDeath(LivingEntity entity, DamageSource source, float amount) {
        if (entity instanceof Villager villager) {
            SimpleContainer inventory = villager.getInventory();
            List<ItemStack> captured = new ArrayList<>();
            for (int i = 0; i < inventory.getContainerSize(); i++) {
                ItemStack stack = inventory.getItem(i);
                if (!stack.isEmpty()) {
                    captured.add(stack.copy());
                }
            }
            pendingCapture.put(villager.getUUID(), captured);
        }
        return true;
    }

    private void onAfterDeath(LivingEntity entity, DamageSource source) {
        if (!(entity instanceof Villager villager) || !(entity.level() instanceof ServerLevel level)) {
            return;
        }
        List<ItemStack> belongings = pendingCapture.remove(villager.getUUID());
        handleDeath(level, villager, belongings == null ? List.of() : belongings);
    }

    private void handleDeath(ServerLevel level, Villager villager, List<ItemStack> belongings) {
        String name = villager.getName().getString();

        holdGrief(level, villager);

        DuskPairing pairing = duskPairing.get();
        if (pairing != null) {
            pairing.bodyTokenFor(villager.getUUID()).ifPresentOrElse(token -> {
                tokens.retire(token);
                LOGGER.info("[death] retired body token {} for {} (never reissued)", token, name);
            }, () -> LOGGER.warn("[death] no body token on record for {}; nothing retired", name));
        }

        BlockPos gravePos = placeGrave(level, villager, name);
        String graveBody = tokens.issue(TokenRegistry.TokenType.BODY, "grave of " + name);
        String gravePlaceToken = tokens.issue(TokenRegistry.TokenType.PLACE, name + "'s grave");
        placeBelongings(level, gravePos, belongings);

        Thing graveThing = Thing.of(null, "k:grave", List.of("readable"), name + "'s grave");
        Thing belongingsThing = Thing.of(null, "k:storage", List.of("storage"), name + "'s things");
        LOGGER.info("[death] {} died; grave body={} place={} at {}; belongings thing={} ({} item(s))",
            name, graveBody, gravePlaceToken, gravePos, belongingsThing, belongings.size());
        LOGGER.info("[death] grave sighting content: {}", graveThing);

        GraveBoardEntry posting = new GraveBoardEntry(name, new Place(gravePlaceToken, name + "'s grave"));
        // TASK-0020 T006: closes V5's decoupling seam — this posting rides the real board
        // (dev.kithcraft.mod.board.Board#postNotice) instead of only being logged.
        boardData.board().postNotice((String) posting.content().get("text"));
        boardData.markDirty();
        LOGGER.info("[death] board posting composed: {}", posting.content());
    }

    /** T007 (card AC #4): find the site (bounded, deterministic search) and place a named
     * marker there — a vanilla oak sign, no new block type. */
    private BlockPos placeGrave(ServerLevel level, Villager villager, String name) {
        GravePlacement.Pos deathPos = toPos(villager.blockPosition());
        GravePlacement.Pos sitePos = GravePlacement.findGraveSite(deathPos, p -> isSafeBuildable(level, toBlockPos(p)));
        BlockPos pos = toBlockPos(sitePos);

        level.setBlockAndUpdate(pos, Blocks.OAK_SIGN.defaultBlockState());
        if (level.getBlockEntity(pos) instanceof SignBlockEntity sign) {
            sign.updateText(t -> t.setMessage(0, Component.literal("Here lies"))
                .setMessage(1, Component.literal(name)), true);
        }
        return pos;
    }

    /** T007 (card AC #5): the pre-captured hidden inventory, placed as an ordinary storage
     * container beside the grave. An empty capture still places an empty chest (deviation:
     * "empty or omitted" per the spec's edge case — this mod chooses "empty, not omitted", so
     * the grave's world footprint is uniform regardless of what the villager was carrying). */
    private void placeBelongings(ServerLevel level, BlockPos gravePos, List<ItemStack> belongings) {
        BlockPos bundlePos = gravePos.offset(1, 0, 0);
        level.setBlockAndUpdate(bundlePos, Blocks.CHEST.defaultBlockState());
        if (level.getBlockEntity(bundlePos) instanceof ChestBlockEntity chest) {
            for (int i = 0; i < belongings.size() && i < chest.getContainerSize(); i++) {
                chest.setItem(i, belongings.get(i));
            }
        }
    }

    /** T008 (FR-006, card AC #7): R-4 found no natural POI re-claim lag at all — the grief
     * period has to actively re-claim the ticket {@code releasePoi} just freed. {@code
     * Villager.releasePoi} never erases the {@code HOME}/{@code JOB_SITE} brain memory it
     * reads (only the ticket), so both are still readable here. */
    private void holdGrief(ServerLevel level, Villager villager) {
        PoiManager poi = level.getPoiManager();
        long now = level.getGameTime();
        for (MemoryModuleType<GlobalPos> type : GRIEF_POI_MEMORIES) {
            villager.getBrain().getMemory(type).ifPresent(globalPos -> {
                BlockPos pos = globalPos.pos();
                Optional<BlockPos> claimed = poi.take(t -> true, (t, p) -> p.equals(pos), pos, 0);
                claimed.ifPresent(p -> {
                    GriefPeriod.Hold hold = GriefPeriod.start(villager.getName().getString() + "'s " + type, now);
                    holds.add(new PendingHold(p, hold));
                    LOGGER.info("[death] grief hold started: {} at {} until tick {}", hold.referent(), p,
                        hold.heldUntilTick());
                });
            });
        }
    }

    /** Called once per server tick: releases any grief hold whose period has elapsed. */
    public void tick(ServerLevel level, long worldTime) {
        holds.removeIf(h -> {
            if (h.hold().isHeld(worldTime)) {
                return false;
            }
            level.getPoiManager().release(h.pos());
            LOGGER.info("[death] grief hold released: {} at {} (tick {})", h.hold().referent(), h.pos(), worldTime);
            return true;
        });
    }

    private static boolean isSafeBuildable(ServerLevel level, BlockPos pos) {
        var below = level.getBlockState(pos.below());
        var here = level.getBlockState(pos);
        return below.isSolid() && below.getFluidState().isEmpty()
            && here.isAir() && here.getFluidState().isEmpty();
    }

    private static GravePlacement.Pos toPos(BlockPos pos) {
        return new GravePlacement.Pos(pos.getX(), pos.getY(), pos.getZ());
    }

    private static BlockPos toBlockPos(GravePlacement.Pos pos) {
        return new BlockPos(pos.x(), pos.y(), pos.z());
    }
}
