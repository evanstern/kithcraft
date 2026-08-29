package dev.kithcraft.mod.build;

import java.util.LinkedHashMap;
import java.util.Locale;
import java.util.Map;
import java.util.Optional;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * T007 (spec.md US3 AC #3: "a posting the engine cannot interpret still supports
 * claim-and-decline cleanly"): a tiny deterministic parser, not a language model — the
 * earliest recognized keyword in the text wins per field, a missing keyword takes the
 * generous default, and only genuinely empty text is unparseable ({@link Optional#empty()}).
 * A gibberish-but-nonblank posting still parses (to the defaults); whether the MIND then
 * claims or declines it is M5's own deliberation, not a judgment this parser makes.
 *
 * <p>ponytail: substring keyword matching over a handful of words, one number, one clamp.
 * "A little hut, not a tower" parses as HUT (earlier mention), not an error — good enough for
 * the demo's fidelity floor, not a real NLU pass. Ceiling: exactly what US3 AC #3 asks for;
 * upgrade path is a real parser/grammar if the demo ever needs one, which per
 * kithcraft-brief.md it deliberately does not.
 */
public final class BlueprintParser {
    private BlueprintParser() {
    }

    private static final int DEFAULT_SIZE = 5;
    private static final int MIN_SIZE = 2;
    private static final int MAX_SIZE = 10;

    private static final Map<String, Blueprint.Shape> SHAPE_KEYWORDS = new LinkedHashMap<>();
    private static final Map<String, Blueprint.Material> MATERIAL_KEYWORDS = new LinkedHashMap<>();

    static {
        SHAPE_KEYWORDS.put("wall", Blueprint.Shape.WALL);
        SHAPE_KEYWORDS.put("hut", Blueprint.Shape.HUT);
        SHAPE_KEYWORDS.put("house", Blueprint.Shape.HUT);
        SHAPE_KEYWORDS.put("tower", Blueprint.Shape.TOWER);
        SHAPE_KEYWORDS.put("pen", Blueprint.Shape.PEN);
        SHAPE_KEYWORDS.put("fence", Blueprint.Shape.PEN);

        MATERIAL_KEYWORDS.put("wood", Blueprint.Material.WOOD);
        MATERIAL_KEYWORDS.put("plank", Blueprint.Material.WOOD);
        MATERIAL_KEYWORDS.put("stone", Blueprint.Material.STONE);
        MATERIAL_KEYWORDS.put("cobble", Blueprint.Material.STONE);
        MATERIAL_KEYWORDS.put("brick", Blueprint.Material.BRICK);
        MATERIAL_KEYWORDS.put("dirt", Blueprint.Material.DIRT);
        MATERIAL_KEYWORDS.put("mud", Blueprint.Material.DIRT);
    }

    private static final Pattern NUMBER = Pattern.compile("\\d+");

    /** {@code null}/blank text is the only unparseable case — a posting with nothing to read
     * at all. Anything else always produces a {@link Blueprint}: shape defaults to WALL,
     * material to WOOD, size to {@value #DEFAULT_SIZE} (clamped {@value #MIN_SIZE}-
     * {@value #MAX_SIZE} regardless of what the text names, keeping every build tractable
     * within a work period's tick budget). */
    public static Optional<Blueprint> parse(String text) {
        if (text == null || text.isBlank()) {
            return Optional.empty();
        }
        String lower = text.toLowerCase(Locale.ROOT);
        Blueprint.Shape shape = earliestMatch(lower, SHAPE_KEYWORDS).orElse(Blueprint.Shape.WALL);
        Blueprint.Material material = earliestMatch(lower, MATERIAL_KEYWORDS).orElse(Blueprint.Material.WOOD);
        int size = extractSize(lower).orElse(DEFAULT_SIZE);
        return Optional.of(new Blueprint(shape, Math.max(MIN_SIZE, Math.min(MAX_SIZE, size)), material));
    }

    /** The value whose keyword occurs earliest in {@code lower} — "a little hut, not a tower"
     * picks HUT, since "hut" appears first. */
    private static <T> Optional<T> earliestMatch(String lower, Map<String, T> keywords) {
        int bestIndex = Integer.MAX_VALUE;
        T best = null;
        for (Map.Entry<String, T> entry : keywords.entrySet()) {
            int index = lower.indexOf(entry.getKey());
            if (index >= 0 && index < bestIndex) {
                bestIndex = index;
                best = entry.getValue();
            }
        }
        return Optional.ofNullable(best);
    }

    private static Optional<Integer> extractSize(String lower) {
        Matcher m = NUMBER.matcher(lower);
        return m.find() ? Optional.of(Integer.parseInt(m.group())) : Optional.empty();
    }
}
