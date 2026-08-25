# T003 — dev-server load observation

Ran `./gradlew runServer` in `mod/` (with `run/eula.txt` containing `eula=true` — standard
dev-server setup, `run/` is gitignored and not committed). The server started, loaded the
mod, and reached `Done`:

```
[13:17:10] [main/INFO] (FabricLoader/Mixin) SpongePowered MIXIN Subsystem ...
...
	- kithcraft 0.1.0
	- minecraft 26.2
	...
[13:17:15] [main/INFO] (kithcraft) kithcraft mod initialized
[13:17:15] [main/ERROR] (Minecraft) Failed to load properties from file: server.properties
java.nio.file.NoSuchFileException: server.properties
	... (expected on first run — no server.properties exists yet; the server falls back to
	     defaults and creates the file; non-fatal, unrelated to the mod)
[13:17:17] [Server thread/INFO] (Minecraft) Starting minecraft server version 26.2
[13:17:17] [Server thread/INFO] (Minecraft) Starting Minecraft server on *:25565
[13:17:19] [Server thread/INFO] (Minecraft) Preparing spawn area: 100%
[13:17:19] [Server thread/INFO] (Minecraft) Done (1.952s)! For help, type "help"
```

The mod id (`kithcraft 0.1.0`) appears in the loaded-mods list and the mod's own log line
(`kithcraft mod initialized`) fires from `KithcraftMod.onInitialize()` before world
generation — confirms the entrypoint runs (card AC #7 partial: mod loads without error).

**Stopped** via `SIGTERM` to the server process (no interactive console attached to the
background-run process; Gradle reports the resulting non-zero exit as a build failure,
which is expected for a forced stop and not a build-gate concern — `./gradlew build` was
already verified green separately, before this run).

**No client jar**: `mod/build/libs/` contains exactly one artifact, `kithcraft-0.1.0.jar`
(server jar only) — confirms card AC #7 / decision-0002.
