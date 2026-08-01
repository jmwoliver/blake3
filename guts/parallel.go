package guts

// This bounded scheduling design was adapted from goforge.dev/blake3sum
// v1.0.0, Copyright 2026 brain-fuel, MIT license; see LICENSE-GOFORGE.

import (
	"runtime"
	"sync"
	"sync/atomic"
)

func parallelFor(n int, fn func(int)) {
	if n <= 0 {
		return
	}
	workers := runtime.GOMAXPROCS(0)
	if n < 2 || workers <= 1 {
		for i := 0; i < n; i++ {
			fn(i)
		}
		return
	}
	if workers > n {
		workers = n
	}
	var next atomic.Int64
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for {
				i := int(next.Add(1)) - 1
				if i >= n {
					return
				}
				fn(i)
			}
		}()
	}
	wg.Wait()
}
