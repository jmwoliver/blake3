# Third-party notices

This fork starts from `lukechampine.com/blake3` v1.4.1, Copyright 2020 Luke
Champine, under the MIT license in `LICENSE`.

The ARM64 NEON kernels in `guts/compress_arm64.s`, their dispatch integration
in `guts/compress_arm64.go`, and the bounded scheduling design in `parallel.go`
and `guts/parallel.go` were ported or adapted from
`goforge.dev/blake3sum` v1.0.0, Copyright 2026 brain-fuel, under the MIT license
in `LICENSE-GOFORGE`.

Source revisions:

- `lukechampine.com/blake3` v1.4.1: commit
  `dd9ffb94b32f2dd34f8605c8a74a03e05a2d3ae2`
- `goforge.dev/blake3sum` v1.0.0: commit
  `a6d86d69518ab24ca652e8a93f93cf45f175c03f`
