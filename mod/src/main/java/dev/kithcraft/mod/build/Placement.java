package dev.kithcraft.mod.build;

import java.util.ArrayList;
import java.util.List;

/**
 * T007/T008: turns a {@link Blueprint} into a deterministic, ordered list of relative block
 * offsets — the one placement plan {@link BuildEngine} walks one tick's worth of at a time.
 * Pure integer math, no Minecraft types, so the placement ORDER (what card AC #6's
 * interrupt/resume depends on: resuming must place the SAME next block, not a different one)
 * is testable without a world.
 *
 * <p>ponytail: four shapes, all hollow/frame silhouettes — no roof, no floor, no doors or
 * windows. A shape a player can watch rise, not a livable structure. Ceiling: exactly what
 * "built beside the player" (US3) needs; a real structure generator is out of this demo's
 * scope.
 */
public final class Placement {
    private Placement() {
    }

    public record Offset(int x, int y, int z) {
    }

    private static final int WALL_HEIGHT = 3;
    private static final int HUT_HEIGHT = 3;
    private static final int PEN_HEIGHT = 1;

    public static List<Offset> plan(Blueprint blueprint) {
        int size = blueprint.size();
        return switch (blueprint.shape()) {
            case WALL -> wall(size);
            case TOWER -> tower(size);
            case HUT -> hollowSquare(size, HUT_HEIGHT);
            case PEN -> hollowSquare(size, PEN_HEIGHT);
        };
    }

    private static List<Offset> wall(int length) {
        List<Offset> offsets = new ArrayList<>();
        for (int y = 0; y < WALL_HEIGHT; y++) {
            for (int x = 0; x < length; x++) {
                offsets.add(new Offset(x, y, 0));
            }
        }
        return offsets;
    }

    private static List<Offset> tower(int height) {
        List<Offset> offsets = new ArrayList<>();
        for (int y = 0; y < height; y++) {
            offsets.add(new Offset(0, y, 0));
        }
        return offsets;
    }

    /** A hollow perimeter (walls only, no interior fill) of a {@code size}x{@code size}
     * footprint, {@code height} tall — HUT's frame and PEN's low fence are the same shape
     * function at two different heights. */
    private static List<Offset> hollowSquare(int size, int height) {
        List<Offset> offsets = new ArrayList<>();
        for (int y = 0; y < height; y++) {
            for (int x = 0; x < size; x++) {
                for (int z = 0; z < size; z++) {
                    if (x == 0 || z == 0 || x == size - 1 || z == size - 1) {
                        offsets.add(new Offset(x, y, z));
                    }
                }
            }
        }
        return offsets;
    }
}
