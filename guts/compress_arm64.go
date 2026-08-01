//go:build arm64

package guts

// The NEON kernels in compress_arm64.s were ported from
// goforge.dev/blake3sum v1.0.0. Copyright 2026 brain-fuel, MIT license; see
// LICENSE-GOFORGE and THIRD_PARTY_NOTICES.md in the module root.

import (
	"encoding/binary"
	"unsafe"
)

// compress_arm64.s loads Node fields by fixed byte offset. Fail the build if a
// future Node edit changes that ABI without updating the assembly.
var (
	_ [0]byte   = [unsafe.Offsetof(Node{}.CV)]byte{}
	_ [32]byte  = [unsafe.Offsetof(Node{}.Block)]byte{}
	_ [96]byte  = [unsafe.Offsetof(Node{}.Counter)]byte{}
	_ [104]byte = [unsafe.Offsetof(Node{}.BlockLen)]byte{}
	_ [108]byte = [unsafe.Offsetof(Node{}.Flags)]byte{}
	_ [112]byte = [unsafe.Sizeof(Node{})]byte{}
)

// ARMv8-A requires Advanced SIMD, so the NEON path is always available on
// GOARCH=arm64. Keeping this as a variable lets package tests force the
// portable implementation for differential checks.
var haveNEON = true

//go:noescape
func compressNodeNEON(n *Node, out *[16]uint32)

// compressChunksNEON compresses four full chunks in parallel.
//
//go:noescape
func compressChunksNEON(cvs *[4][8]uint32, buf *[4 * ChunkSize]byte, key *[8]uint32, counter uint64, flags uint32)

// CompressBuffer compresses up to MaxSIMD chunks in parallel and returns their
// root node.
func CompressBuffer(buf *[MaxSIMD * ChunkSize]byte, buflen int, key *[8]uint32, counter uint64, flags uint32) Node {
	if buflen <= ChunkSize {
		return CompressChunk(buf[:buflen], key, counter, flags)
	}
	if !haveNEON {
		return compressBufferGeneric(buf, buflen, key, counter, flags)
	}

	var cvs [MaxSIMD][8]uint32
	numChunks := uint64(buflen / ChunkSize)
	var group uint64
	for ; (group+1)*4 <= numChunks; group++ {
		compressChunksNEON(
			(*[4][8]uint32)(unsafe.Pointer(&cvs[group*4])),
			(*[4 * ChunkSize]byte)(unsafe.Pointer(&buf[group*4*ChunkSize])),
			key,
			counter+group*4,
			flags,
		)
	}
	for i := group * 4; i < numChunks; i++ {
		cvs[i] = ChainingValue(CompressChunk(buf[i*ChunkSize:(i+1)*ChunkSize], key, counter+i, flags))
	}
	numCVs := numChunks
	if buflen%ChunkSize != 0 {
		cvs[numCVs] = ChainingValue(CompressChunk(buf[numCVs*ChunkSize:buflen], key, counter+numCVs, flags))
		numCVs++
	}
	return mergeSubtrees(&cvs, numCVs, key, flags)
}

// CompressChunk compresses a single chunk, returning its final uncompressed
// node.
func CompressChunk(chunk []byte, key *[8]uint32, counter uint64, flags uint32) Node {
	n := Node{
		CV:       *key,
		Counter:  counter,
		BlockLen: BlockSize,
		Flags:    flags | FlagChunkStart,
	}
	var block [BlockSize]byte
	for len(chunk) > BlockSize {
		copy(block[:], chunk)
		chunk = chunk[BlockSize:]
		n.Block = BytesToWords(block)
		n.CV = ChainingValue(n)
		n.Flags &^= FlagChunkStart
	}
	block = [BlockSize]byte{}
	n.BlockLen = uint32(copy(block[:], chunk))
	n.Block = BytesToWords(block)
	n.Flags |= FlagChunkEnd
	return n
}

// CompressBlocks compresses MaxSIMD copies of n with successive counters.
func CompressBlocks(out *[MaxSIMD * BlockSize]byte, n Node) {
	var outs [MaxSIMD][BlockSize]byte
	compressBlocksGeneric(&outs, n)
	for i := range outs {
		copy(out[i*BlockSize:], outs[i][:])
	}
}

func mergeSubtrees(cvs *[MaxSIMD][8]uint32, numCVs uint64, key *[8]uint32, flags uint32) Node {
	return mergeSubtreesGeneric(cvs, numCVs, key, flags)
}

// BytesToWords converts 64 little-endian bytes to 16 words.
func BytesToWords(block [BlockSize]byte) (words [16]uint32) {
	for i := range words {
		words[i] = binary.LittleEndian.Uint32(block[4*i:])
	}
	return words
}

// WordsToBytes converts 16 words to 64 little-endian bytes.
func WordsToBytes(words [16]uint32) (block [BlockSize]byte) {
	for i, word := range words {
		binary.LittleEndian.PutUint32(block[4*i:], word)
	}
	return block
}
