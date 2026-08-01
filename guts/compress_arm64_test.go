//go:build arm64

package guts

import (
	"math/rand"
	"testing"
)

func TestCompressNodeNEONMatchesGeneric(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	flags := []uint32{
		0,
		FlagChunkStart,
		FlagChunkEnd,
		FlagChunkStart | FlagChunkEnd | FlagRoot,
		FlagParent,
		FlagKeyedHash | FlagChunkStart | FlagChunkEnd,
		FlagDeriveKeyContext | FlagRoot,
		FlagDeriveKeyMaterial | FlagRoot,
	}
	blockLens := []uint32{0, 1, 31, 63, 64}

	for i := 0; i < 1_000; i++ {
		var n Node
		for j := range n.CV {
			n.CV[j] = rng.Uint32()
		}
		for j := range n.Block {
			n.Block[j] = rng.Uint32()
		}
		n.Counter = rng.Uint64()
		n.BlockLen = blockLens[i%len(blockLens)]
		n.Flags = flags[i%len(flags)]

		want := CompressNode(n)
		var got [16]uint32
		compressNodeNEON(&n, &got)
		if got != want {
			t.Fatalf("case %d: NEON compression mismatch\nwant %08x\n got %08x", i, want, got)
		}
	}
}

func TestCompressChunksNEONMatchesGeneric(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	var buf [4 * ChunkSize]byte
	if _, err := rng.Read(buf[:]); err != nil {
		t.Fatal(err)
	}
	var key [8]uint32
	for i := range key {
		key[i] = rng.Uint32()
	}
	const counter = uint64(0xfedcba9876543210)
	const flags = uint32(FlagKeyedHash)

	var got [4][8]uint32
	compressChunksNEON(&got, &buf, &key, counter, flags)

	old := haveNEON
	haveNEON = false
	t.Cleanup(func() { haveNEON = old })
	for i := range got {
		chunk := buf[i*ChunkSize : (i+1)*ChunkSize]
		want := ChainingValue(CompressChunk(chunk, &key, counter+uint64(i), flags))
		if got[i] != want {
			t.Fatalf("chunk %d: NEON compression mismatch\nwant %08x\n got %08x", i, want, got[i])
		}
	}
}

func TestCompressBufferNEONMatchesGeneric(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	var buf [MaxSIMD * ChunkSize]byte
	if _, err := rng.Read(buf[:]); err != nil {
		t.Fatal(err)
	}
	var key [8]uint32
	for i := range key {
		key[i] = rng.Uint32()
	}
	lengths := []int{
		0, 1, BlockSize - 1, BlockSize, BlockSize + 1,
		ChunkSize - 1, ChunkSize, ChunkSize + 1,
		2 * ChunkSize, 4 * ChunkSize, 4*ChunkSize + 1,
		8 * ChunkSize, 15*ChunkSize + 17, MaxSIMD * ChunkSize,
	}
	flags := []uint32{0, FlagKeyedHash, FlagDeriveKeyMaterial}

	for _, length := range lengths {
		for _, flag := range flags {
			haveNEON = true
			got := CompressBuffer(&buf, length, &key, 37, flag)
			haveNEON = false
			want := CompressBuffer(&buf, length, &key, 37, flag)
			if got != want {
				t.Fatalf("length %d flags %d: NEON buffer mismatch\nwant %+v\n got %+v", length, flag, want, got)
			}
		}
	}
	haveNEON = true
}
