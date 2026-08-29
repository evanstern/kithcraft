package dev.kithcraft.mod.build;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** T007/T008: {@link Placement#plan} is pure, deterministic offset math — the ordering {@link
 * BuildEngine}'s tick-budget stepping (and therefore interrupt/resume, card AC #6) depends
 * on. */
class PlacementTest {

    @Test
    void wallIsALineOfLengthTimesHeightBlocks() {
        List<Placement.Offset> plan = Placement.plan(new Blueprint(Blueprint.Shape.WALL, 5, Blueprint.Material.WOOD));
        assertEquals(15, plan.size(), "5 long x 3 tall");
        assertTrue(plan.contains(new Placement.Offset(0, 0, 0)));
        assertTrue(plan.contains(new Placement.Offset(4, 2, 0)));
    }

    @Test
    void towerIsAVerticalColumnOfHeightBlocks() {
        List<Placement.Offset> plan = Placement.plan(new Blueprint(Blueprint.Shape.TOWER, 4, Blueprint.Material.STONE));
        assertEquals(4, plan.size());
        assertEquals(new Placement.Offset(0, 3, 0), plan.get(3));
    }

    @Test
    void hutIsAHollowPerimeterNotAFilledBox() {
        List<Placement.Offset> plan = Placement.plan(new Blueprint(Blueprint.Shape.HUT, 3, Blueprint.Material.BRICK));
        // 3x3 footprint, 3 tall: perimeter is 8 cells/layer (center excluded) * 3 layers.
        assertEquals(24, plan.size());
        assertTrue(plan.stream().noneMatch(o -> o.x() == 1 && o.z() == 1), "center is never placed (hollow)");
    }

    @Test
    void penIsAOneBlockTallHollowPerimeter() {
        List<Placement.Offset> plan = Placement.plan(new Blueprint(Blueprint.Shape.PEN, 3, Blueprint.Material.DIRT));
        assertEquals(8, plan.size());
        assertTrue(plan.stream().allMatch(o -> o.y() == 0));
    }

    @Test
    void planIsDeterministicAcrossCalls() {
        Blueprint blueprint = new Blueprint(Blueprint.Shape.WALL, 5, Blueprint.Material.WOOD);
        assertEquals(Placement.plan(blueprint), Placement.plan(blueprint));
    }
}
