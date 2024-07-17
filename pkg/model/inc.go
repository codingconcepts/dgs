package model

import "sync/atomic"

type Sequence func() int64

// Inc returns a thread-safe sequence generator, starting from
// a given number.
func Inc(start int64) Sequence {
	// Reduce start by one so that the first increment results in
	// the expected value.
	start--

	return func() int64 {
		return atomic.AddInt64(&start, 1)
	}
}
