package dev.kithcraft.mod.death;

import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.regex.Pattern;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;

/**
 * T006 (card ACs #10, #11): the two spell-breakers death-mechanics.md §1/§4 rules against stay
 * structurally absent from the mod's own source — a grep-style regression guard over the real
 * source tree, same philosophy V2/V3 used for their no-Mixin/no-salience claims (the mod
 * source is the ground truth, not a doc's promise), so a future addition trips this test rather
 * than landing silently.
 *
 * <p>TASK-0020 T006 (spec.md User Story 2 AC #2, card AC #9) reuses this same style for a
 * different mod: no force-claim path exists anywhere in the mod source.
 */
class StructuralAbsenceTest {

    /** Resolves mod/src/main/java robustly: prefer the path Gradle passes explicitly, else
     * walk up from the working directory (same pattern as VectorSuiteTest.vectorsDir()). */
    private static Path mainSrcDir() {
        String configured = System.getProperty("kithcraft.mainSrcDir");
        if (configured != null && Files.isDirectory(Path.of(configured))) {
            return Path.of(configured);
        }
        for (Path dir = Path.of("").toAbsolutePath(); dir != null; dir = dir.getParent()) {
            Path candidate = dir.resolve("mod/src/main/java");
            if (Files.isDirectory(candidate)) {
                return candidate;
            }
        }
        throw new IllegalStateException("could not locate mod/src/main/java from " + Path.of("").toAbsolutePath());
    }

    private static String allMainSource() throws IOException {
        try (Stream<Path> files = Files.walk(mainSrcDir())) {
            StringBuilder all = new StringBuilder();
            for (Path p : files.filter(f -> f.toString().endsWith(".java")).toList()) {
                all.append(Files.readString(p)).append('\n');
            }
            return all.toString();
        }
    }

    /** Card AC #10: no self-preservation surface — death-mechanics.md §4 notes hunger isn't
     * even a death vector on Villager, so nothing should exist to feed, escort, or keep
     * vigilance over one. */
    @Test
    void noSelfPreservationSurface() throws IOException {
        Pattern forbidden = Pattern.compile("(?i)\\bfeed(ing)?\\b|\\bescort(ing)?\\b|\\bvigilan(t|ce)\\b");
        assertFalse(forbidden.matcher(allMainSource()).find(),
            "card AC #10: no self-preservation surface (feeding/escort/vigilance) may exist in mod source");
    }

    /** Card AC #11: no engine guardrail on friendly fire — death-mechanics.md §1 rules a
     * direct player kill ADMIT, no guardrail added; the player's capacity for real harm is a
     * deliberate design choice, not an omission to patch later. */
    @Test
    void noFriendlyFireGuardrail() throws IOException {
        Pattern forbidden = Pattern.compile("(?i)friendly.?fire");
        assertFalse(forbidden.matcher(allMainSource()).find(),
            "card AC #11: no friendly-fire guardrail may exist in mod source");
    }

    /** TASK-0020 T006 (card AC #9): checked as call-site cardinality rather than a
     * forbidden-word grep — {@code dev.kithcraft.mod.board.Board#tryClaim} (the only place a
     * claim is ever registered) is called from exactly one place in the whole mod source:
     * {@code BoardClaims}, resolving a real claim intent (T004). A second call site anywhere
     * — a command handler, an admin API, anything player-triggered — would be a force-claim
     * path and trips this test rather than landing silently. */
    @Test
    void noForceClaimPath() throws IOException {
        long callSites = Pattern.compile("\\.tryClaim\\(").matcher(allMainSource()).results().count();
        assertEquals(1L, callSites,
            "card AC #9: Board.tryClaim must be called from exactly one place in the mod — "
                + "the engine's own claim-intent resolution (BoardClaims), never a forced/player path");
    }
}
