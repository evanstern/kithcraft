package dev.kithcraft.mod.build;

/**
 * T007 (spec.md US3 AC #3, plan.md design decision 4): the parsed shape of a blueprint
 * posting — the demo's declared fidelity floor. Free text becomes exactly this three-field
 * guess; nothing richer (multi-axis dimensions, door/window placement, a roof on every shape)
 * is modeled.
 *
 * <p>ponytail: three fields, one size number, four shapes, four materials — the ceiling is
 * "recognizable enough to build something the player can watch rise beside them," not a
 * construction game (kithcraft-brief.md's loneliness-cure constraint). Upgrade path: a real
 * schematic/NBT-structure format, if the demo ever needs richer builds — deliberately not
 * reached for here.
 */
public record Blueprint(Shape shape, int size, Material material) {
    public enum Shape { WALL, HUT, TOWER, PEN }

    public enum Material { WOOD, STONE, BRICK, DIRT }
}
