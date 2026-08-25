package dev.kithcraft.mod.wire;

import org.junit.jupiter.api.Test;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.channels.Channels;
import java.nio.channels.ReadableByteChannel;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

/** Frame-layer rules no vector fixture can carry, because a malformed frame is not a frame
 * (seam-wire-v0.md §2.3/§2.5) — every case here must be connection-fatal. */
class FrameCodecTest {

    private static ReadableByteChannel channel(byte[] bytes) {
        return Channels.newChannel(new ByteArrayInputStream(bytes));
    }

    @Test
    void zeroLengthIsFatal() {
        assertThrows(FrameCodec.FramingException.class,
            () -> FrameCodec.readFrame(channel(new byte[] {0, 0, 0, 0})));
    }

    @Test
    void shortHeaderIsFatal() {
        assertThrows(FrameCodec.FramingException.class,
            () -> FrameCodec.readFrame(channel(new byte[] {0, 0, 1})));
    }

    @Test
    void truncatedBodyIsFatal() {
        assertThrows(FrameCodec.FramingException.class,
            () -> FrameCodec.readFrame(channel(new byte[] {0, 0, 0, 8, '{', '}'})));
    }

    @Test
    void overCapIsFatal() {
        // length word = 0x00200001, just over the 1 MiB cap.
        assertThrows(FrameCodec.FramingException.class,
            () -> FrameCodec.readFrame(channel(new byte[] {0, 0x20, 0, 1})));
    }

    @Test
    void cleanEofBetweenFramesIsOrderly() throws IOException {
        assertNull(FrameCodec.readFrame(channel(new byte[0])));
    }

    @Test
    void backToBackFramesDecodeIndependently() throws IOException {
        ByteArrayOutputStream out1 = new ByteArrayOutputStream();
        FrameCodec.writeFrame(Channels.newChannel(out1), CanonicalJson.encode(Map.of("a", 1L)));
        ByteArrayOutputStream out2 = new ByteArrayOutputStream();
        FrameCodec.writeFrame(Channels.newChannel(out2), CanonicalJson.encode(Map.of("b", 2L)));

        ByteArrayOutputStream stream = new ByteArrayOutputStream();
        stream.writeBytes(out1.toByteArray());
        stream.writeBytes(out2.toByteArray());

        ReadableByteChannel in = channel(stream.toByteArray());
        byte[] first = FrameCodec.readFrame(in);
        byte[] second = FrameCodec.readFrame(in);
        assertEquals(Map.of("a", 1L), CanonicalJson.decode(first));
        assertEquals(Map.of("b", 2L), CanonicalJson.decode(second));
    }

    @Test
    void writeThenReadRoundTrips() throws IOException {
        Map<String, Object> value = Map.of("hello", "world", "n", 42L);
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        FrameCodec.writeFrame(Channels.newChannel(out), CanonicalJson.encode(value));

        byte[] body = FrameCodec.readFrame(channel(out.toByteArray()));
        assertArrayEquals(CanonicalJson.encode(value), body);
        assertEquals(value, CanonicalJson.decode(body));
    }
}
