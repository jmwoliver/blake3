package guts

// This bounded scheduling design was adapted from goforge.dev/blake3sum
// v1.0.0, Copyright 2026 brain-fuel, MIT license; see LICENSE-GOFORGE.

import (
	"runtime"
	"sync"
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
	var wg sync.WaitGroup
	if n <= workers {
		wg.Add(n)
		for i := range n {
			go func(i int) {
				defer wg.Done()
				fn(i)
			}(i)
		}
	} else {
		wg.Add(workers)
		for worker := range workers {
			go func(worker int) {
				defer wg.Done()
				for i := worker; i < n; i += workers {
					fn(i)
				}
			}(worker)
		}
	}
	wg.Wait()
}
