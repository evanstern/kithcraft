package dev.kithcraft.mod.build;

import org.junit.jupiter.api.Test;

import java.util.Optional;

import static org.junit.jupiter.api.Assertions.assertEquals;

/** T007 (spec.md US3 AC #3): the fidelity-floor parser's vocabulary, its generous defaults,
 * and the one case (blank text) it genuinely can't interpret. */
class BlueprintParserTest {

    @Test
    void blankOrNullTextIsUnparseable() {
        assertEquals(Optional.empty(), BlueprintParser.parse(null));
        assertEquals(Optional.empty(), BlueprintParser.parse("   "));
    }

    @Test
    void freeTextWithNoRecognizedWordsStillParsesToGenerousDefaults() {
        Blueprint blueprint = BlueprintParser.parse("something, anything, whatever you like").orElseThrow();
        assertEquals(Blueprint.Shape.WALL, blueprint.shape());
        assertEquals(Blueprint.Material.WOOD, blueprint.material());
        assertEquals(5, blueprint.size());
    }

    @Test
    void recognizesShapeAndMaterialKeywordsCaseInsensitively() {
        Blueprint blueprint = BlueprintParser.parse("Build me a STONE Tower, ten blocks tall").orElseThrow();
        assertEquals(Blueprint.Shape.TOWER, blueprint.shape());
        assertEquals(Blueprint.Material.STONE, blueprint.material());
    }

    @Test
    void sizeClampsToAReasonableRange() {
        assertEquals(10, BlueprintParser.parse("a tower 500 blocks tall").orElseThrow().size());
        assertEquals(2, BlueprintParser.parse("a tower 0 blocks tall").orElseThrow().size());
    }

    @Test
    void earliestKeywordOccurrenceWins() {
        Blueprint blueprint = BlueprintParser.parse("a little hut, not a tower").orElseThrow();
        assertEquals(Blueprint.Shape.HUT, blueprint.shape(), "\"hut\" appears before \"tower\"");
    }

    @Test
    void recognizesThePenAndFenceSynonyms() {
        assertEquals(Blueprint.Shape.PEN, BlueprintParser.parse("pen for the sheep").orElseThrow().shape());
        assertEquals(Blueprint.Shape.PEN, BlueprintParser.parse("a fence around back").orElseThrow().shape());
    }
}
