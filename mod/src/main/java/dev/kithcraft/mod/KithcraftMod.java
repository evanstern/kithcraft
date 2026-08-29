package dev.kithcraft.mod;

import dev.kithcraft.mod.brain.DuskPairing;
import dev.kithcraft.mod.cast.Cast;
import dev.kithcraft.mod.cast.CastData;
import dev.kithcraft.mod.cast.CastSeeder;
import dev.kithcraft.mod.live.BodySession;
import dev.kithcraft.mod.live.LiveDeathHandling;
import dev.kithcraft.mod.tokens.TokenRegistryData;
import net.fabricmc.api.ModInitializer;
import net.fabricmc.fabric.api.event.lifecycle.v1.ServerEntityEvents;
import net.fabricmc.fabric.api.event.lifecycle.v1.ServerLifecycleEvents;
import net.fabricmc.fabric.api.event.lifecycle.v1.ServerTickEvents;
import net.minecraft.core.BlockPos;
import net.minecraft.server.MinecraftServer;
import net.minecraft.server.level.ServerLevel;
import net.minecraft.world.entity.npc.villager.Villager;
import net.minecraft.world.phys.AABB;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.List;
import java.util.Set;
import java.util.stream.Collectors;

/**
 * Server-side entrypoint (decision-0002: no client jar). T011's live wiring: on each server
 * tick, attach a {@link BodySession} to the first {@code Villager} found once the mind
 * daemon's socket is dialable, then drive that session's tick every tick after.
 *
 * <p>ponytail: single-body attach, retried at most once/sec while unattached — a real
 * multi-body attach loop (spawn events, per-body retry backoff) is V3's job; this is a
 * dev-server proof, not the production session manager.
 */
public class KithcraftMod implements ModInitializer {
	public static final Logger LOGGER = LoggerFactory.getLogger("kithcraft");

	private BodySession attached;
	// Deliberately not Long.MIN_VALUE: `worldTime - lastAttachAttempt` below would overflow
	// (worldTime is a small non-negative game-tick count) and wrap to a huge negative value,
	// which is always < 20 — silently skipping every attach attempt forever.
	private long lastAttachAttempt = -1000L;
	private DuskPairing duskPairing;
	private TokenRegistryData pendingTokens;
	private BlockPos pendingOrigin;
	private long lastDuskSetupAttempt = -1000L;
	private LiveDeathHandling deathHandling;

	@Override
	public void onInitialize() {
		LOGGER.info("kithcraft mod initialized");
		ServerLifecycleEvents.SERVER_STARTED.register(this::onServerStarted);
		ServerTickEvents.END_SERVER_TICK.register(this::onServerTick);
		// T011 diagnostic: ground truth for whether/where an entity actually loads,
		// independent of the getEntitiesOfClass scan below (which is returning 0 even after
		// a confirmed "Summoned new Villager" console line — this narrows whether the entity
		// truly never loads or whether the scan itself is the bug).
		ServerEntityEvents.ENTITY_LOAD.register((entity, level) ->
			LOGGER.info("[live] ENTITY_LOAD: {} at {}", entity.getClass().getName(), entity.blockPosition()));
	}

	/** T002 (card AC #6): seed the three-member cast once per world, at the world spawn.
	 * Idempotent via {@link CastData}'s persisted flag — a restart re-runs this but does not
	 * re-spawn. */
	private void onServerStarted(MinecraftServer server) {
		var level = server.overworld();
		CastData cast = level.getDataStorage().computeIfAbsent(CastData.TYPE);
		BlockPos origin = level.getRespawnData().pos();
		CastSeeder.seedIfNeeded(level, cast, origin);

		// T007 (FR-005, R-7): the shared gathering place + each matched cast member's body
		// token. NOT built here: at SERVER_STARTED, persisted villagers have not finished
		// loading yet (their ENTITY_LOAD fires a few ticks later — confirmed live, T009), so
		// findCastVillagers() would return empty and DuskPairing would never have any seats.
		// Deferred to onServerTick below, retried until the cast is actually found.
		pendingTokens = level.getDataStorage().computeIfAbsent(TokenRegistryData.TYPE);
		pendingOrigin = origin;
		// T011 dev-server proof (card AC #9): every token that still resolves at boot, so a
		// retired token's absence here after a restart is directly observable in the log.
		LOGGER.info("[live] token registry live entries at boot: {}", pendingTokens.liveEntries());

		// T007-T009: registers the Fabric API death hooks (no new Mixin — see
		// LiveDeathHandling's class doc). One registration per server start; the supplier
		// reads this.duskPairing lazily since it isn't set up until a later tick.
		deathHandling = LiveDeathHandling.register(pendingTokens, () -> duskPairing);
	}

	private static List<Villager> findCastVillagers(ServerLevel level) {
		AABB huge = new AABB(-1000, -256, -1000, 1000, 256, 1000);
		Set<String> castNames = Cast.MEMBERS.stream().map(Cast.Member::name).collect(Collectors.toSet());
		return level.getEntitiesOfClass(Villager.class, huge).stream()
			.filter(v -> castNames.contains(v.getName().getString()))
			.toList();
	}

	private void onServerTick(MinecraftServer server) {
		try {
			onServerTickUnsafe(server);
		} catch (Throwable t) {
			// T011 diagnostic: Fabric's event dispatch does not always surface a listener
			// exception loudly, so this pass logs everything at ERROR rather than risk a
			// silent no-op (found live: an earlier long-overflow bug in the throttle below
			// produced exactly this symptom — no attach, no log, no error).
			LOGGER.error("[live] onServerTick threw", t);
		}
	}

	private void onServerTickUnsafe(MinecraftServer server) throws IOException {
		long worldTime = server.overworld().getGameTime();
		if (duskPairing == null && pendingTokens != null && worldTime - lastDuskSetupAttempt >= 20) {
			lastDuskSetupAttempt = worldTime;
			var castVillagers = findCastVillagers(server.overworld());
			if (!castVillagers.isEmpty()) {
				// T011 finding (full-cycle-observation.md, 2026-08-27): re-cover the cast's
				// actual current chunk footprint every boot, not just where CastSeeder
				// force-loaded at first spawn — see CastSeeder.keepChunksLoaded's own javadoc
				// for why a stale/too-tight footprint permanently freezes a villager's Brain
				// without ever fully unloading it.
				CastSeeder.keepChunksLoaded(server.overworld(), castVillagers);
				duskPairing = DuskPairing.setUp(server.overworld(), pendingTokens, pendingOrigin, castVillagers);
				pendingTokens = null;
			}
		}
		if (duskPairing != null) {
			duskPairing.tick(worldTime);
		}
		if (deathHandling != null) {
			deathHandling.tick(server.overworld(), worldTime);
		}
		if (attached == null) {
			if (worldTime - lastAttachAttempt < 20) {
				return;
			}
			lastAttachAttempt = worldTime;
			AABB huge = new AABB(-1000, -256, -1000, 1000, 256, 1000);
			var villagers = server.overworld().getEntitiesOfClass(Villager.class, huge);
			var everything = server.overworld().getEntitiesOfClass(net.minecraft.world.entity.Entity.class, huge);
			LOGGER.info("[live] attach scan at tick {}: {} villager(s), {} total entities: {}", worldTime,
				villagers.size(), everything.size(),
				everything.stream().map(e -> e.getClass().getName()).distinct().toList());
			Villager villager = villagers.stream().findFirst().orElse(null);
			if (villager == null) {
				return;
			}
			try {
				attached = BodySession.open(server, villager);
				LOGGER.info("[live] attached to villager {}", villager.getStringUUID());
			} catch (IOException e) {
				LOGGER.warn("[live] mind dial failed, will retry: {}", e.toString());
			}
			return;
		}
		try {
			attached.tick(worldTime);
		} catch (IOException e) {
			LOGGER.warn("[live] session tick failed, dropping session: {}", e.toString());
			attached = null;
		}
	}
}
