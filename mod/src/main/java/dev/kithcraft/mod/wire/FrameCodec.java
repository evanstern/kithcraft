package dev.kithcraft.mod.wire;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.channels.ReadableByteChannel;
import java.nio.channels.WritableByteChannel;

/**
 * Length-prefixed framing (seam-wire-v0.md §2): a 4-byte big-endian length word, then
 * exactly that many bytes of UTF-8 JSON body. Framing failures are always connection-fatal
 * (§2.3/§2.5) — once frame boundaries are in doubt there is no resynchronization, so every
 * method here throws {@link FramingException} rather than attempting recovery.
 */
public final class FrameCodec {
    /** §2.5: maximum frame body, enforced before allocation. */
    public static final int MAX_FRAME_BODY = 1 << 20;

    private FrameCodec() {}

    /** Any framing-layer failure. The caller MUST treat the connection as failed (§2.3). */
    public static final class FramingException extends IOException {
        public FramingException(String message) {
            super(message);
        }
    }

    /**
     * Reads exactly one frame's body, looping on short reads (§2.3).
     *
     * @return the frame body, or {@code null} on an orderly disconnect (EOF between
     *     frames, before any byte of the next length word was read).
     * @throws FramingException on a truncated frame, a zero length, or a length exceeding
     *     the cap.
     */
    public static byte[] readFrame(ReadableByteChannel channel) throws IOException {
        ByteBuffer header = ByteBuffer.allocate(4);
        if (!readFully(channel, header, true)) {
            return null; // clean EOF between frames
        }
        header.flip();
        long length = header.getInt() & 0xFFFFFFFFL;
        if (length == 0) {
            throw new FramingException("malformed frame: length 0 (an empty body is not a JSON value)");
        }
        if (length > MAX_FRAME_BODY) {
            throw new FramingException("frame body " + length + " exceeds the 1 MiB cap (§2.5)");
        }
        ByteBuffer body = ByteBuffer.allocate((int) length);
        readFully(channel, body, false);
        return body.array();
    }

    /** Writes one frame — length word and body together — as a single buffer (§4.2: never
     * split a frame across writes). */
    public static void writeFrame(WritableByteChannel channel, byte[] body) throws IOException {
        if (body.length > MAX_FRAME_BODY) {
            throw new FramingException("frame body " + body.length + " exceeds the 1 MiB cap (§2.5)");
        }
        ByteBuffer buf = ByteBuffer.allocate(4 + body.length);
        buf.putInt(body.length);
        buf.put(body);
        buf.flip();
        while (buf.hasRemaining()) {
            int n = channel.write(buf);
            if (n < 0) {
                throw new FramingException("write failed mid-frame");
            }
        }
    }

    /**
     * Reads until {@code buf} is full or the channel is exhausted.
     *
     * @param allowLeadingEof if true, an EOF with nothing yet read is a clean disconnect
     *     ({@code false} is returned); any other EOF is a truncated frame.
     * @return true once {@code buf} is full; false only for the clean-disconnect case above.
     */
    private static boolean readFully(ReadableByteChannel channel, ByteBuffer buf, boolean allowLeadingEof)
            throws IOException {
        int total = 0;
        while (buf.hasRemaining()) {
            int n = channel.read(buf);
            if (n == -1) {
                if (allowLeadingEof && total == 0) {
                    return false;
                }
                throw new FramingException("truncated frame: stream ended after " + total + " byte(s)");
            }
            total += n;
        }
        return true;
    }
}
