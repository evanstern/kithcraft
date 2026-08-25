package dev.kithcraft.mod;

import net.fabricmc.api.ModInitializer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/** Server-side entrypoint (decision-0002: no client jar). Phase 1 skeleton only — no wire,
 * manifest, or registry yet (Phases 2-3). */
public class KithcraftMod implements ModInitializer {
	public static final Logger LOGGER = LoggerFactory.getLogger("kithcraft");

	@Override
	public void onInitialize() {
		LOGGER.info("kithcraft mod initialized");
	}
}
