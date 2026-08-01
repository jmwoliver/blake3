# blake3

This is a performance-focused fork of `lukechampine.com/blake3` v1.4.1. It
retains the upstream AVX2, AVX-512, portable implementation, and public API,
and adds an ARM64 four-chunk NEON backend, bounded tree-hash scheduling, and
zero-allocation keyed one-shot helpers.

```sh
go get github.com/jacobwoliver/blake3
```

`blake3` implements the [BLAKE3 cryptographic hash function](https://github.com/BLAKE3-team/BLAKE3).
This implementation aims to be performant without sacrificing (too much)
readability, in the hopes of eventually landing in `x/crypto`.

Architecture dispatch is internal to the module:

- AMD64 uses the upstream AVX-512 or AVX2 assembly when supported.
- ARM64 uses four-lane NEON for groups of independent full chunks and the
  portable single-node compressor for small messages and Merkle parents.
- Other architectures use portable Go.

`NewSerial` provides the same streaming API without internal worker
goroutines for applications that already batch independent messages.

The ARM64 kernels were ported from `goforge.dev/blake3sum` v1.0.0. See
`THIRD_PARTY_NOTICES.md`, `LICENSE`, and `LICENSE-GOFORGE` for provenance and
license terms.

## Historical upstream benchmarks

Tested on a 2020 MacBook Air (i5-7600K @ 3.80GHz). Benchmarks will improve as
soon as I get access to a beefier AVX-512 machine. :wink:

### AVX-512

```
BenchmarkSum256/64           120 ns/op       533.00 MB/s
BenchmarkSum256/1024        2229 ns/op       459.36 MB/s
BenchmarkSum256/65536      16245 ns/op      4034.11 MB/s
BenchmarkWrite               245 ns/op      4177.38 MB/s
BenchmarkXOF                 246 ns/op      4159.30 MB/s
```

### AVX2

```
BenchmarkSum256/64           120 ns/op       533.00 MB/s
BenchmarkSum256/1024        2229 ns/op       459.36 MB/s
BenchmarkSum256/65536      31137 ns/op      2104.76 MB/s
BenchmarkWrite               487 ns/op      2103.12 MB/s
BenchmarkXOF                 329 ns/op      3111.27 MB/s
```

### Pure Go

```
BenchmarkSum256/64           120 ns/op       533.00 MB/s
BenchmarkSum256/1024        2229 ns/op       459.36 MB/s
BenchmarkSum256/65536     133505 ns/op       490.89 MB/s
BenchmarkWrite              2022 ns/op       506.36 MB/s
BenchmarkXOF                1914 ns/op       534.98 MB/s
```

## Shortcomings

There is no assembly routine for single-block compressions. This is most
noticeable for ~1KB inputs.

Each assembly routine inlines all 7 rounds, causing thousands of lines of
duplicated code. Ideally the routines could be merged such that only a single
routine is generated for AVX-512 and AVX2, without sacrificing too much
performance.
