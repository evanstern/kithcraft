package dev.kithcraft.mod.death;

import dev.kithcraft.mod.percept.ChangeReports;
import dev.kithcraft.mod.percept.PerceptEmitter;
import dev.kithcraft.mod.percept.Provenance;
import dev.kithcraft.mod.percept.Sightings;
import dev.kithcraft.mod.percept.Thing;
import dev.kithcraft.mod.session.Handshake;
import dev.kithcraft.mod.wire.CanonicalJson;
import dev.kithcraft.mod.wire.FrameCodec;
import dev.kithcraft.mod.wire.WireClient;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.net.StandardProtocolFamily;
import java.net.UnixDomainSocketAddress;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.SocketChannel;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * T010 (FR-008, card AC #8): death rides the ordinary percept channels only — no death percept
 * type exists (structural absence, mirroring {@link StructuralAbsenceTest}'s philosophy), a
 * witness sees an ordinary {@code sighting}, and an absent villager gets §4.10's restricted
 * {@code change_report} (change:"gone") plus a {@code sighting} of the grave on return. The
 * third test pushes that composed content over a real (loopback UDS) wire session, the same
 * stub-mind harness {@code HandshakeWireClientTest} established — proving delivery, not just
 * composition.
 */
class DeathPerceptChannelTest {

    private static Thing dyingVillager() {
        return Thing.of(null, "k:person", List.of(), "a villager");
    }

    private static Thing grave() {
        return Thing.of("th-grave-1", "k:grave", List.of("readable"), "Tam's grave");
    }

    /** No death percept type exists anywhere in the closed vocabulary a vendor may declare —
     * death is not a wire concept, only an ordinary sighting/change_report of ordinary things. */
    @Test
    @SuppressWarnings("unchecked")
    void noDeathPerceptTypeExistsInTheClosedVocabulary() {
        List<String> perceptTypes = (List<String>) Handshake.MANIFEST.get("percept_types");
        assertFalse(perceptTypes.stream().anyMatch(t -> t.toLowerCase(java.util.Locale.ROOT).contains("death")),
            "card AC #8: no death-shaped percept type may be declared — " + perceptTypes);
    }

    /** A witnessing villager's percept is an ordinary sighting: thing/doing/origin:"saw" —
     * composed with the same {@link Sightings}/{@link PerceptEmitter} machinery any other
     * sighting uses, no death-specific composer. */
    @Test
    void witnessGetsAnOrdinarySightingOfTheDyingVillager() throws Exception {
        PerceptEmitter emitter = new PerceptEmitter(m -> {}, "s-death-1");
        Provenance sawIt = Provenance.direct(Provenance.Origin.SAW, 500L, 500L);
        Map<String, Object> content = Sightings.sightingContent(dyingVillager(), "near", "dying");

        Map<String, Object> envelope = emitter.compose(
            "b-witness", "sighting", PerceptEmitter.Urgency.NOTABLE, sawIt, null, content, 500L);

        assertEquals("sighting", ((Map<String, Object>) envelope.get("payload")).get("percept_type"));
        assertEquals("saw", sawIt.toPayload().get("origin"));
        assertEquals("dying", content.get("doing"));
    }

    /** §4.10's restriction resolves to exactly the absent body (no acting villager caused this
     * death — a zombie did, which never holds a body token); on return that body gets a
     * change_report (change:"gone") for the dead villager plus an ordinary sighting of the new
     * grave thing — card AC #8's second half. */
    @Test
    @SuppressWarnings("unchecked")
    void absentVillagerGetsChangeReportGoneThenAGraveSighting() throws Exception {
        Set<String> candidates = Set.of("b-witness", "b-absent");
        Set<String> recipients = ChangeReports.recipients(candidates, null, Set.of("b-witness"));
        assertEquals(Set.of("b-absent"), recipients, "only the absent body may receive the change_report");

        PerceptEmitter emitter = new PerceptEmitter(m -> {}, "s-death-1");
        Map<String, Object> changeContent = ChangeReports.content(ChangeReports.Change.GONE, dyingVillager());
        Provenance changeProv = ChangeReports.provenance(1000L, 1000L);
        Map<String, Object> changeEnvelope = emitter.compose(
            "b-absent", "change_report", PerceptEmitter.Urgency.NOTABLE, changeProv, null, changeContent, 1000L);

        assertEquals("gone", changeContent.get("change"));
        assertEquals("change_report",
            ((Map<String, Object>) changeEnvelope.get("payload")).get("percept_type"));

        Map<String, Object> graveContent = Sightings.sightingContent(grave(), "here", null);
        Map<String, Object> graveEnvelope = emitter.compose(
            "b-absent", "sighting", PerceptEmitter.Urgency.BACKGROUND, changeProv, null, graveContent, 1000L);

        assertEquals("sighting",
            ((Map<String, Object>) graveEnvelope.get("payload")).get("percept_type"));
        assertEquals(grave().toPayload(), graveContent.get("thing"));
    }

    // --- live wire delivery, HandshakeWireClientTest's stub-mind harness reused ---

    private Path socketPath;
    private ServerSocketChannel listener;

    @BeforeEach
    void setUp() throws IOException {
        socketPath = Path.of("/tmp", "kc-death-test-" + System.nanoTime() + ".sock");
        Files.deleteIfExists(socketPath);
        listener = ServerSocketChannel.open(StandardProtocolFamily.UNIX);
        listener.bind(UnixDomainSocketAddress.of(socketPath));
    }

    @AfterEach
    void tearDown() throws IOException {
        listener.close();
        Files.deleteIfExists(socketPath);
    }

    /** Pushes the witness sighting and the absent villager's change_report + grave sighting
     * over a real {@link WireClient} connection to a stub mind listener — proving the composed
     * content actually crosses the wire, not just that it composes correctly in memory. */
    @Test
    void composedDeathContentCrossesALiveWireSession() throws Exception {
        CompletableFuture<List<Object>> mindReceived = new CompletableFuture<>();
        CompletableFuture<Void> mindDone = CompletableFuture.runAsync(() -> {
            try (SocketChannel conn = listener.accept()) {
                List<Object> received = new java.util.ArrayList<>();
                for (int i = 0; i < 3; i++) {
                    received.add(CanonicalJson.decode(FrameCodec.readFrame(conn)));
                }
                mindReceived.complete(received);
            } catch (IOException e) {
                mindReceived.completeExceptionally(e);
            }
        });

        try (WireClient client = new WireClient(socketPath)) {
            client.dial();
            assertTrue(client.isConnected());

            PerceptEmitter emitter = PerceptEmitter.overWireClient(client, "s-death-live");
            Provenance sawIt = Provenance.direct(Provenance.Origin.SAW, 500L, 500L);
            emitter.emit("b-witness", "sighting", PerceptEmitter.Urgency.NOTABLE, sawIt, null,
                Sightings.sightingContent(dyingVillager(), "near", "dying"), 500L);

            Provenance changeProv = ChangeReports.provenance(1000L, 1000L);
            emitter.emit("b-absent", "change_report", PerceptEmitter.Urgency.NOTABLE, changeProv, null,
                ChangeReports.content(ChangeReports.Change.GONE, dyingVillager()), 1000L);
            emitter.emit("b-absent", "sighting", PerceptEmitter.Urgency.BACKGROUND, changeProv, null,
                Sightings.sightingContent(grave(), "here", null), 1000L);

            List<Object> received = mindReceived.get(5, TimeUnit.SECONDS);
            assertEquals(3, received.size());
            @SuppressWarnings("unchecked")
            Map<String, Object> sightingEnvelope = (Map<String, Object>) received.get(0);
            @SuppressWarnings("unchecked")
            Map<String, Object> changeEnvelope = (Map<String, Object>) received.get(1);
            @SuppressWarnings("unchecked")
            Map<String, Object> graveEnvelope = (Map<String, Object>) received.get(2);

            assertEquals("b-witness", sightingEnvelope.get("body"));
            assertEquals("b-absent", changeEnvelope.get("body"));
            assertEquals("b-absent", graveEnvelope.get("body"));
        }

        mindDone.get(5, TimeUnit.SECONDS);
    }
}
