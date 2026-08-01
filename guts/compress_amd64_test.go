//go:build amd64

package guts

import (
	"math/rand"
	"testing"
	"unsafe"
)

func TestCompressBufferAVX2MatchesGeneric(t *testing.T) {
	if !haveAVX2 {
		t.Skip("AVX2 is not available on this host")
	}
	buf, key := amd64TestInput(11)
	for _, length := range []int{
		2 * ChunkSize, 2*ChunkSize + 1, 7*ChunkSize + 13,
		8 * ChunkSize, 9*ChunkSize + 17, 15 * ChunkSize,
		MaxSIMD * ChunkSize,
	} {
		for _, flags := range []uint32{0, FlagKeyedHash, FlagDeriveKeyMaterial} {
			got := compressBufferAVX2(&buf, length, &key, 37, flags)
			want := compressBufferGeneric(&buf, length, &key, 37, flags)
			if got != want {
				t.Fatalf("length %d flags %d: AVX2 buffer mismatch\nwant %+v\n got %+v", length, flags, want, got)
			}
		}
	}
}

func TestCompressBufferAVX512MatchesGeneric(t *testing.T) {
	if !haveAVX512 {
		t.Skip("AVX-512 is not available on this host")
	}
	buf, key := amd64TestInput(12)
	for _, length := range []int{
		2 * ChunkSize, 2*ChunkSize + 1, 7*ChunkSize + 13,
		8 * ChunkSize, 9*ChunkSize + 17, 15 * ChunkSize,
		MaxSIMD * ChunkSize,
	} {
		for _, flags := range []uint32{0, FlagKeyedHash, FlagDeriveKeyMaterial} {
			got := compressBufferAVX512(&buf, length, &key, 41, flags)
			want := compressBufferGeneric(&buf, length, &key, 41, flags)
			if got != want {
				t.Fatalf("length %d flags %d: AVX-512 buffer mismatch\nwant %+v\n got %+v", length, flags, want, got)
			}
		}
	}
}

func TestCompressBlocksAVX2MatchesGeneric(t *testing.T) {
	if !haveAVX2 {
		t.Skip("AVX2 is not available on this host")
	}
	n := amd64TestNode(13)
	var got [MaxSIMD * BlockSize]byte
	halves := (*[2][8 * BlockSize]byte)(unsafe.Pointer(&got))
	compressBlocksAVX2(&halves[0], &n.Block, &n.CV, n.Counter, n.BlockLen, n.Flags)
	compressBlocksAVX2(&halves[1], &n.Block, &n.CV, n.Counter+8, n.BlockLen, n.Flags)

	var want [MaxSIMD][BlockSize]byte
	compressBlocksGeneric(&want, n)
	if got != *(*[MaxSIMD * BlockSize]byte)(unsafe.Pointer(&want)) {
		t.Fatal("AVX2 output-block compression mismatch")
	}
}

func TestCompressBlocksAVX512MatchesGeneric(t *testing.T) {
	if !haveAVX512 {
		t.Skip("AVX-512 is not available on this host")
	}
	n := amd64TestNode(14)
	var got [MaxSIMD * BlockSize]byte
	compressBlocksAVX512(&got, &n.Block, &n.CV, n.Counter, n.BlockLen, n.Flags)

	var want [MaxSIMD][BlockSize]byte
	compressBlocksGeneric(&want, n)
	if got != *(*[MaxSIMD * BlockSize]byte)(unsafe.Pointer(&want)) {
		t.Fatal("AVX-512 output-block compression mismatch")
	}
}

func TestCompressParentsAVX2MatchesGeneric(t *testing.T) {
	if !haveAVX2 {
		t.Skip("AVX2 is not available on this host")
	}
	_, key := amd64TestInput(15)
	rng := rand.New(rand.NewSource(16))
	var input [MaxSIMD][8]uint32
	for i := range input {
		for j := range input[i] {
			input[i][j] = rng.Uint32()
		}
	}
	for numCVs := uint64(2); numCVs <= MaxSIMD; numCVs++ {
		gotCVs, wantCVs := input, input
		got := mergeSubtrees(&gotCVs, numCVs, &key, FlagKeyedHash)
		want := mergeSubtreesGeneric(&wantCVs, numCVs, &key, FlagKeyedHash)
		if got != want {
			t.Fatalf("%d chaining values: AVX2 parent compression mismatch\nwant %+v\n got %+v", numCVs, want, got)
		}
	}
}

func amd64TestInput(seed int64) (buf [MaxSIMD * ChunkSize]byte, key [8]uint32) {
	rng := rand.New(rand.NewSource(seed))
	if _, err := rng.Read(buf[:]); err != nil {
		panic(err)
	}
	for i := range key {
		key[i] = rng.Uint32()
	}
	return
}

func amd64TestNode(seed int64) (n Node) {
	rng := rand.New(rand.NewSource(seed))
	for i := range n.CV {
		n.CV[i] = rng.Uint32()
	}
	for i := range n.Block {
		n.Block[i] = rng.Uint32()
	}
	n.Counter = rng.Uint64()
	n.BlockLen = BlockSize - 1
	n.Flags = FlagKeyedHash | FlagRoot
	return
}
