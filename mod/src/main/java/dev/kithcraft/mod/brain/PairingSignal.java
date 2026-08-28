package dev.kithcraft.mod.brain;

import dev.kithcraft.mod.percept.Sightings;
import dev.kithcraft.mod.percept.Thing;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.OptionalDouble;

/**
 * T008 (FR-005, R-7): the pairing signal's pure timing math and content, plus the per-pair
 * "already fired this approach" bookkeeping so a multi-tick approach signals exactly once
 * instead of on every tick it happens to sit inside the lead window. No Minecraft types —
 * {@link DuskPairing} is the live driver that feeds this each tick's remaining-distance/speed
 * estimate for a candidate pair.
 *
 * <p><b>Why {@code sighting}, not a new percept type</b>: plan.md's own Constitution Check
 * already settled this — "the pairing signal reaches the mind via existing percept types
 * (sighting/self_state at the gathering place)" — no protocol extension (FR-008). A {@code
 * sighting} is the least-contorted existing fit: {@link Thing#body} already carries one
 * opaque body token (AR-3), and the envelope's own {@code place} field (never coordinates,
 * AR-4) carries the gathering place — exactly "the pair's body tokens + place token in seam
 * terms" FR-005 asks for, with no new wire shape. It is composed twice per firing pair, once
 * per perceiver, each one sighting the OTHER member approaching the shared place — the
 * perceiving member's own body token is the envelope's {@code body} field ({@link
 * dev.kithcraft.mod.percept.PerceptEmitter#compose}), not part of this content.
 */
public final class PairingSignal {

    /** R-7's fixed lead time: the signal fires this many seconds before predicted arrival —
     * never on or after arrival. */
    public static final double LEAD_SECONDS = 10.0;

    private final Map<String, Boolean> firedForPair = new HashMap<>();

    /** The pure math: seconds until arrival at the given speed, or empty when pathing has
     * failed/stalled (R-7's no-fire edge case — {@code remainingBlocks} empty) or speed is
     * non-positive (a stuck/stationary villager can never arrive). */
    public static OptionalDouble predictedArrivalSeconds(OptionalDouble remainingBlocks, double blocksPerSecond) {
        if (remainingBlocks.isEmpty() || blocksPerSecond <= 0) {
            return OptionalDouble.empty();
        }
        return OptionalDouble.of(remainingBlocks.getAsDouble() / blocksPerSecond);
    }

    /** A pair arrives when its SLOWER member arrives — combines two members' individual ETAs
     * into the pair's ETA. Empty if either member's own ETA is empty (one pathing failure
     * fails the whole pair's approach this dusk, per the spec edge case). */
    public static OptionalDouble combine(OptionalDouble etaA, OptionalDouble etaB) {
        if (etaA.isEmpty() || etaB.isEmpty()) {
            return OptionalDouble.empty();
        }
        return OptionalDouble.of(Math.max(etaA.getAsDouble(), etaB.getAsDouble()));
    }

    /** @return true the FIRST tick the pair's predicted arrival is within the lead window
     * (and arrival hasn't already happened); false on every later tick of the same approach,
     * and false whenever {@code pairEtaSeconds} is empty (pathing failed for either member).
     * Call {@link #reset} once pathing fails or the pair actually arrives, so a fresh
     * approach can signal again. */
    public boolean shouldFireNow(String pairKey, OptionalDouble pairEtaSeconds) {
        Objects.requireNonNull(pairKey, "pairKey");
        if (pairEtaSeconds.isEmpty()) {
            return false;
        }
        double eta = pairEtaSeconds.getAsDouble();
        if (eta <= 0 || eta > LEAD_SECONDS) {
            return false;
        }
        return firedForPair.putIfAbsent(pairKey, Boolean.TRUE) == null;
    }

    /** Clears a pair's "already fired" bookkeeping. */
    public void reset(String pairKey) {
        firedForPair.remove(pairKey);
    }

    /** The {@code sighting} content one pair member perceives about the OTHER, ~10s before
     * arrival (see class javadoc for why {@code sighting} is the chosen vehicle). {@code
     * otherBody} rides on {@link Thing#body} (AR-3); the place token is the caller's job to
     * pass as the envelope's {@code place} (never coordinates). */
    public static Map<String, Object> content(String otherBody, String otherName) {
        Objects.requireNonNull(otherBody, "otherBody");
        Objects.requireNonNull(otherName, "otherName");
        Thing other = new Thing(null, "k:person", List.of(), otherName, otherBody, 1);
        return Sightings.sightingContent(other, "near", "walking to the gathering place");
    }
}
