package dev.kithcraft.mod.wire;

import java.io.IOException;
import java.net.StandardProtocolFamily;
import java.net.UnixDomainSocketAddress;
import java.nio.channels.SocketChannel;
import java.nio.file.Path;

/**
 * Dials the mind daemon's Unix domain socket. Per decision-0004 / seam-wire-v0.md §1.2,
 * the vendor always dials and never listens.
 *
 * <p>Re-dial is bounded exponential backoff (§1.5's stated default: 250ms doubling, capped
 * at 10s) — the shape is normative, the constants are configuration. Full disconnected-body
 * behaviour (bodies kept alive engine-side while unattended) is V3's; this is the minimal
 * client the handshake needs.
 */
public final class WireClient implements AutoCloseable {
    private static final long INITIAL_BACKOFF_MS = 250;
    private static final long MAX_BACKOFF_MS = 10_000;

    private final Path socketPath;
    private SocketChannel channel;

    public WireClient(Path socketPath) {
        this.socketPath = socketPath;
    }

    /** One dial attempt. Throws on failure — callers wanting retry use {@link #dialWithBackoff}. */
    public void dial() throws IOException {
        SocketChannel c = SocketChannel.open(StandardProtocolFamily.UNIX);
        c.connect(UnixDomainSocketAddress.of(socketPath));
        channel = c;
    }

    /**
     * Dials with bounded exponential backoff between attempts.
     *
     * @param maxAttempts total dial attempts before giving up; {@code <= 0} means retry
     *     indefinitely (the daemon may be down for a while — T-4 means that is survivable).
     */
    public void dialWithBackoff(int maxAttempts) throws IOException, InterruptedException {
        long backoff = INITIAL_BACKOFF_MS;
        int attempt = 0;
        while (true) {
            attempt++;
            try {
                dial();
                return;
            } catch (IOException e) {
                if (maxAttempts > 0 && attempt >= maxAttempts) {
                    throw e;
                }
                Thread.sleep(backoff);
                backoff = Math.min(backoff * 2, MAX_BACKOFF_MS);
            }
        }
    }

    /** Encodes and sends one message as a single frame. */
    public void send(Object message) throws IOException {
        FrameCodec.writeFrame(channel, CanonicalJson.encode(message));
    }

    /** Reads and decodes one frame. @return the decoded message, or {@code null} on an
     * orderly disconnect. */
    public Object receive() throws IOException {
        byte[] body = FrameCodec.readFrame(channel);
        return body == null ? null : CanonicalJson.decode(body);
    }

    public boolean isConnected() {
        return channel != null && channel.isConnected();
    }

    @Override
    public void close() throws IOException {
        if (channel != null) {
            channel.close();
        }
    }
}
