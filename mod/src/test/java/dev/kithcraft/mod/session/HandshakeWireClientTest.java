package dev.kithcraft.mod.session;

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
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * Proves {@link WireClient} + {@link Handshake} against a stub mind listener — a plain Java
 * UNIX-socket {@link ServerSocketChannel} accepting one connection, no daemon dependency
 * (TASK-0008's daemon is a separate lane; per spec.md's Assumptions, the two must not
 * depend on each other). Dial, send {@code session_open}, receive a scripted ack, clean
 * close.
 */
class HandshakeWireClientTest {

    private Path socketPath;
    private ServerSocketChannel listener;

    @BeforeEach
    void setUp() throws IOException {
        // Short and under sun_path's ~103-byte limit (seam-wire-v0.md §1.1) — deliberately
        // not java.io.tmpdir, whose per-process macOS path can itself exceed that budget.
        socketPath = Path.of("/tmp", "kc-test-" + System.nanoTime() + ".sock");
        Files.deleteIfExists(socketPath);
        listener = ServerSocketChannel.open(StandardProtocolFamily.UNIX);
        listener.bind(UnixDomainSocketAddress.of(socketPath));
    }

    @AfterEach
    void tearDown() throws IOException {
        listener.close();
        Files.deleteIfExists(socketPath);
    }

    @Test
    void dialsSendsSessionOpenAndReceivesScriptedAck() throws Exception {
        Object scriptedAck = Map.of("ack", true, "session", "s-test-1");

        // The stub mind: accept once, read one frame, hand it back to the test thread,
        // then reply with the scripted ack.
        CompletableFuture<Object> mindReceived = new CompletableFuture<>();
        CompletableFuture<Void> mindDone = CompletableFuture.runAsync(() -> {
            try (SocketChannel conn = listener.accept()) {
                byte[] body = FrameCodec.readFrame(conn);
                mindReceived.complete(CanonicalJson.decode(body));
                FrameCodec.writeFrame(conn, CanonicalJson.encode(scriptedAck));
            } catch (IOException e) {
                mindReceived.completeExceptionally(e);
            }
        });

        Map<String, Object> sessionOpen = Handshake.buildSessionOpen(
            "s-test-1", "b-test-1", 1000L, Continuity.firstSession());

        try (WireClient client = new WireClient(socketPath)) {
            client.dial();
            assertTrue(client.isConnected());

            client.send(sessionOpen);
            Object receivedByMind = mindReceived.get(5, TimeUnit.SECONDS);
            assertEquals(sessionOpen, receivedByMind, "mind must decode exactly what the vendor sent");

            Object ack = client.receive();
            assertEquals(scriptedAck, ack, "vendor must decode exactly the mind's scripted ack");
        }

        mindDone.get(5, TimeUnit.SECONDS); // propagate any stub-side exception
    }
}
