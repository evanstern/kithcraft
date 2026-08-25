# Version re-verification (T001)

**Evidence rule**: every claim below carries a URL and the date it was checked. Checked
2026-08-25.

## Chosen toolchain

| Component | Version | Source |
|---|---|---|
| Minecraft | 26.2 (current stable Java Edition) | https://www.mcrpg.com/articles/latest-minecraft-update — "26.2 is the current stable Java Edition release," released June 16, 2026; confirmed 26.3 is still snapshot-only (not stable) at https://minecraft.wiki/w/Java_Edition_26.3 (accessed 2026-08-25: "upcoming game drop... set to release in the third quarter of 2026") |
| Fabric Loader | 0.19.3 | https://maven.fabricmc.net/net/fabricmc/fabric-loader/maven-metadata.xml (top version, accessed 2026-08-25); confirmed as current stable by https://fabricmc.net/2026/06/15/262.html ("Players should install the latest stable version of Fabric Loader (currently 0.19.3)") |
| Fabric API | 0.158.0+26.2 | https://github.com/FabricMC/fabric-api/releases (tagged "Latest", accessed 2026-08-25); also top `+26.2` entry in https://maven.fabricmc.net/net/fabricmc/fabric-api/fabric-api/maven-metadata.xml. Note: the bare maven `<latest>` field pointed at 0.158.2+26.3, a **pre-release** build for the not-yet-stable 26.3 snapshot line — not used. |
| Fabric Loom | 1.17.19 | https://maven.fabricmc.net/fabric-loom/fabric-loom.gradle.plugin/maven-metadata.xml (accessed 2026-08-25) — latest non-alpha patch in the 1.17 line fabricmc.net's 26.2 post recommends ("Developers should use Loom 1.17," https://fabricmc.net/2026/06/15/262.html). Pinned to a concrete release rather than the official template's floating `1.17-SNAPSHOT` (https://github.com/FabricMC/fabric-example-mod/blob/26.2/gradle.properties) for reproducibility. |
| Gradle | 9.5.1 | https://raw.githubusercontent.com/FabricMC/fabric-example-mod/26.2/gradle/wrapper/gradle-wrapper.properties (accessed 2026-08-25); matches fabricmc.net's 26.2 post ("Gradle 9.5.1 (at time of writing)"). |
| Java (source/target) | 25 (release flag), JDK 26.0.2 host | https://docs.fabricmc.net/develop/getting-started/setting-up (accessed 2026-08-25: "To develop mods for Minecraft 26.1, you will need JDK 25"); host JDK 26.0.2 satisfies the ">=25" floor (verified TASK-0007). |
| Mappings | **none — Mojang official names, unobfuscated** | https://fabricmc.net/2026/03/14/261.html (accessed 2026-08-25): "26.1 is the first version of Minecraft to not be obfuscated... Yarn is no longer officially supported by Fabric." https://docs.fabricmc.net/develop/porting/mappings/: "Minecraft 26.1 is unobfuscated and includes parameter names, so there is no need for any obfuscation mappings." |

## villager-brain-api.md re-check (flag, not re-derive)

`docs/wiki/villager-brain-api.md` records Yarn-mapped symbol names (`Brain<E>`,
`VillagerEntity`, `MemoryModuleType`, `Schedule`, `Activity`, `SecondaryPointsOfInterestSensor`,
etc.) checked at **yarn-1.21.3+build.1**. The target for this task, MC 26.2, is two major
version lines past Yarn's last supported release (1.21.11 — see
https://fabricmc.net/2025/12/05/12111.html, accessed 2026-08-25: "the plan is to stop
updating Yarn and Intermediary after 1.21.11"). Yarn does not exist for 26.2 at all; MC 26.2
uses Mojang's official (unobfuscated) names directly.

The Fabric API porting guide confirms names moved, not just mapping *style*, in the
26.1 transition: "API names have been updated to match the official names where
applicable... not backwards compatible" (https://fabricmc.net/2026/03/14/261.html,
accessed 2026-08-25), citing e.g. `ItemGroupEvents` → `CreativeModeTabEvents`. There is no
evidence in this pass that the specific brain-API surface (`Brain`, `Schedule`, `Activity`,
`MemoryModuleType`, sensors) was one of the renamed names, but the vanilla-engine classes
(as opposed to Fabric API's own types) are exactly the ones affected by the
obfuscation-name change itself (Yarn `VillagerEntity` → official `Villager`, Yarn
`Brain<E>` → likely official `Brain<E>` unchanged in shape but confirmed only under
Mojang mappings, etc.).

**Flagged, not resolved here** (full re-verification is V3's — entity/brain implementation
work): every symbol in `docs/wiki/villager-brain-api.md`'s "How it works" section needs a
fresh check against Mojang's official 26.2 mappings before any brain/schedule/Mixin code is
written. This is a toolchain-level fact (mappings regime changed under the target), not a
symbol-by-symbol re-derivation, which stays out of this phase's scope per the task
instructions.

## Routing A-2 daylight arithmetic flag

Per the spec's edge cases and the task's routing note, "verify daylight arithmetic against
target version" is **recorded for V3, not resolved here** — no action taken in this phase
beyond carrying the flag forward.

## Dev-server load observation (T003)

See `specs/009-fabric-mod-skeleton/research/runserver-observation.md`.

## JSON library choice and C-1..C-10 emit check (T004)

**Chosen: Gson** (`com.google.gson`) — already on the mod's compile classpath transitively
through the `minecraft` configuration (Minecraft itself bundles and uses Gson; confirmed by a
clean `./gradlew build` compiling `import com.google.gson.stream.JsonReader` with no new
dependency line added to `build.gradle`). Per the ladder's rung 4 (already-installed dependency),
no new dependency was introduced.

**C-1..C-10 emit-check verdict: custom writer needed — Gson cannot be made to emit canonical
form as a thin wrapper.** Two structural gaps, neither configurable away:

- **C-3** (ascending-by-UTF-8-byte key order): Gson's `JsonObject`/`JsonWriter` preserve
  insertion order only; there is no sorted-key mode.
- **C-6** (numbers are integers only, signed-64-bit range, refuse otherwise): Gson's writer
  and reader are permissive about numeric form (doubles, exponents, out-of-range values) with
  no built-in mode to reject them.

Per the operator's 2026-08-22 ruling's carve-out ("the canonical writer stays custom only if
the chosen library provably cannot emit C-1..C-10 canonical form"), `mod/.../wire/CanonicalJson.java`
implements:

- **decode**: uses Gson's `JsonReader` (streaming tokenizer) to parse — the ruling's "library
  for parsing is expected either way" — with the wire's stricter rules layered on top (C-4
  duplicate-key refusal, C-6 integer-only/range check, C-8 lone-surrogate refusal, C-1
  BOM/strict-UTF-8 check), since Gson's default parsing is lenient about all four.
- **encode**: hand-written (does not use Gson's `JsonWriter`), sorting keys and emitting
  integers/escapes/literals per C-1..C-10 directly.

Verified against all 17 `seam/vectors/` fixtures via the Gradle test suite
(`mod/src/test/java/dev/kithcraft/mod/wire/VectorSuiteTest.java`): every vector's canonical
re-encode is byte-identical to its pinned `frame_hex`. This is the record T010 (Phase 4) will
consume when replacing `seam/java-roundtrip`'s hand-rolled parser with a library-based harness
using this same library and the same verdict.
