package dev.kithcraft.mod;

import dev.kithcraft.mod.tokens.TokenRegistry;
import dev.kithcraft.mod.tokens.TokenRegistryData;
import net.fabricmc.api.ModInitializer;
import net.fabricmc.fabric.api.event.lifecycle.v1.ServerLifecycleEvents;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/** Server-side entrypoint (decision-0002: no client jar). No wire, manifest, or intent
 * execution wiring yet (a body session is opened per V2, not here). */
public class KithcraftMod implements ModInitializer {
	public static final Logger LOGGER = LoggerFactory.getLogger("kithcraft");

	@Override
	public void onInitialize() {
		LOGGER.info("kithcraft mod initialized");

		// TEMPORARY skeleton plumbing (T009, card AC #5): seeds a few probe tokens on first
		// run and logs what resolves on every start, so a manual `runServer` stop/restart
		// can be observed to resolve the same tokens to the same referents. Real token
		// issuance is V2's job, wired to actual world events (villager spawn, place
		// discovery, sighting) rather than a fixed dev-server-start probe.
		ServerLifecycleEvents.SERVER_STARTED.register(server -> {
			TokenRegistryData data = server.overworld().getDataStorage().computeIfAbsent(TokenRegistryData.TYPE);
			if (data.liveEntries().isEmpty()) {
				String body = data.issue(TokenRegistry.TokenType.BODY, "dev-token-probe villager");
				String place = data.issue(TokenRegistry.TokenType.PLACE, "dev-token-probe well");
				String thing = data.issue(TokenRegistry.TokenType.THING, "dev-token-probe bed");
				LOGGER.info("[dev-token-probe] first run: issued {}, {}, {} -> {}", body, place, thing, data.liveEntries());
			} else {
				LOGGER.info("[dev-token-probe] restart: resolved {}", data.liveEntries());
			}
		});
	}
}
