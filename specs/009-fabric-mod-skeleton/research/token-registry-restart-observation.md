# Token registry restart observation (T009, card AC #5)

Persistence mechanism: `net.minecraft.world.level.saveddata.SavedData` (MC 26.2's equivalent
of Yarn's `PersistentState` — see versions.md's "PersistentState API name" note), registered
as `dev.kithcraft.mod.tokens.TokenRegistryData.TYPE` and fetched from
`server.overworld().getDataStorage().computeIfAbsent(TYPE)`. `KithcraftMod.onInitialize()`
wires a **temporary, clearly-marked** dev probe on `ServerLifecycleEvents.SERVER_STARTED`:
seeds three tokens (one each of `body`/`place`/`thing`) on first run, and on every start logs
whichever tokens currently resolve. Real token issuance is V2's job, tied to actual world
events — this probe exists only to exercise persistence end to end.

## Run 1 — first start (fresh world)

```
$ ./gradlew runServer
...
[13:40:23] [Server thread/INFO] (Minecraft) Done (2.130s)! For help, type "help"
[13:40:23] [Server thread/INFO] (kithcraft) [dev-token-probe] first run: issued b-0, pl-1, th-2 -> {b-0=dev-token-probe villager, pl-1=dev-token-probe well, th-2=dev-token-probe bed}
```

Stopped with `SIGTERM` (graceful — the vanilla shutdown hook saves the world before exit):

```
[13:40:23] [Server thread/INFO] (Minecraft) Saving chunks for level 'ServerLevel[world]'/minecraft:overworld
...
[13:40:23] [Server thread/INFO] (Minecraft) ThreadedAnvilChunkStorage: All dimensions are saved
```

Confirmed on disk: `run/world/dimensions/minecraft/overworld/data/kithcraft/token_registry.dat`
was written.

## Run 2 — restart against the same world

```
$ ./gradlew runServer
...
[13:41:25] [Server thread/INFO] (Minecraft) Done (0.251s)! For help, type "help"
[13:41:25] [Server thread/INFO] (kithcraft) [dev-token-probe] restart: resolved {b-0=dev-token-probe villager, pl-1=dev-token-probe well, th-2=dev-token-probe bed}
```

## Result

Same three tokens (`b-0`, `pl-1`, `th-2`) resolved to the exact same referents after a full
process restart, and the probe took the "already seeded" branch rather than reissuing —
proving both halves of card AC #5: tokens survive a restart, and issuance does not repeat once
a token exists. This is the dev-server companion to `TokenRegistryTest`'s pure-core
snapshot-round-trip proof (`mod/src/test/java/dev/kithcraft/mod/tokens/TokenRegistryTest.java`).
