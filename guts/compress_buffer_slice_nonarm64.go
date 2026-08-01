//go:build !arm64

package guts

// compressBufferSlice adapts a short slice to the fixed-size bulk interface
// required by the AMD64 and portable implementations. This preserves the
// upstream behavior on non-ARM64 targets.
func compressBufferSlice(buf []byte, key *[8]uint32, counter uint64, flags uint32) Node {
	buflen := len(buf)
	if cap(buf) < MaxSIMD*ChunkSize {
		buf = append(buf, make([]byte, MaxSIMD*ChunkSize-len(buf))...)
	}
	return CompressBuffer((*[MaxSIMD * ChunkSize]byte)(buf[:MaxSIMD*ChunkSize]), buflen, key, counter, flags)
}
