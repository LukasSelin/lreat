package world

import (
	"runtime"
	"sync"
)

// Reading the world on several goroutines at once.
//
// Exactly one goroutine may change a World and several may read one; see the
// remarks at the top of world.go. This is the little that both halves of a
// tick need in order to keep to that - how many goroutines a pass that only
// reads may spread over, and the one way of spreading it.
//
// Every such pass obeys the same rule, and it is the rule that keeps a run
// the same however the goroutines happen to be scheduled: a worker writes
// only where no other worker looks, nothing a worker writes is read until
// every worker has finished, and no worker draws the world's chance. What
// comes out lands in a slice indexed by the work, and the world is changed
// from it afterwards, in one order, on one goroutine. Two passes are spread
// this way - agents deciding, in system.Decide, and the passes over the
// ground, in EachActiveOver - and both are held against a run of the same
// seed that spread over nothing.

// Workers is how many goroutines the read-only passes of a day may spread
// over: the deciding, which is the bulk of a day where there are people,
// and the passes over the ground, which are the bulk of it where there is
// country. Set it to 1 to do everything one at a time.
var Workers = runtime.NumCPU()

// WorkersFor is how many goroutines to spread n pieces of work over: no
// more than there is work, and never none.
func WorkersFor(n int) int {
	k := Workers
	if n < k {
		k = n
	}
	if k < 1 {
		k = 1
	}
	return k
}

// InParallel runs f for every index below n, spread over workers goroutines,
// and returns when the last of them is done. Index i is worked by i%workers,
// so work that lies together is dealt out between the workers rather than
// landing on one of them.
//
// f must not write anything another call to f can see; results belong in a
// slice indexed by i.
func InParallel(n, workers int, f func(i, worker int)) {
	if workers <= 1 {
		for i := 0; i < n; i++ {
			f(i, 0)
		}
		return
	}
	var wg sync.WaitGroup
	for k := 0; k < workers; k++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			for i := k; i < n; i += workers {
				f(i, k)
			}
		}(k)
	}
	wg.Wait()
}
