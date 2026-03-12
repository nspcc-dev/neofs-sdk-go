package pool

import (
	"sync"
	"sync/atomic"
	"time"
)

type (
	slidingWindow struct {
		windowSize int64 // Nanoseconds
		limit      int64

		// Buckets
		prevCount *atomic.Int64
		currCount *atomic.Int64

		windowStart *atomic.Int64 // Nanoseconds

		rotationMu *sync.Mutex
	}
)

func newSlidingWindow(windowSize time.Duration, limit int64) *slidingWindow {
	var windowStart atomic.Int64
	windowStart.Store(time.Now().UnixNano())

	return &slidingWindow{
		windowSize:  int64(windowSize),
		limit:       limit,
		prevCount:   &atomic.Int64{},
		currCount:   &atomic.Int64{},
		windowStart: &windowStart,
		rotationMu:  &sync.Mutex{},
	}
}

// Allow returns true if the errors count is below the threshold.
func (sw *slidingWindow) Allow() bool {
	now := int64(time.Now().UnixNano())
	sw.advance(now)
	sw.currCount.Add(1)

	curr := sw.currCount.Load()
	prev := sw.prevCount.Load()
	start := sw.windowStart.Load()

	elapsed := now - start
	if elapsed >= sw.windowSize {
		return curr < sw.limit
	}

	// weightedCount = curr + (prev * (windowSize - elapsed) / windowSize).
	weightMult := sw.windowSize - elapsed
	weightedPrev := (prev * weightMult) / sw.windowSize

	return (curr + weightedPrev) < sw.limit
}

// advance handles the rotation of windows when time passes.
func (sw *slidingWindow) advance(now int64) {
	start := sw.windowStart.Load()

	// Fast check: Is rotation needed?
	if now-start < sw.windowSize {
		return
	}

	sw.rotationMu.Lock()
	defer sw.rotationMu.Unlock()

	start = sw.windowStart.Load()
	elapsed := now - start
	if elapsed < sw.windowSize {
		return
	}

	if elapsed < sw.windowSize*2 {
		sw.prevCount.Store(sw.currCount.Load())
	} else {
		// Too much time passed, reset prev.
		sw.prevCount.Store(0)
	}

	sw.currCount.Store(0)
	sw.windowStart.Store(now - (now % sw.windowSize))
}

// Current returns current amount of events.
func (sw *slidingWindow) Current() int64 {
	return sw.currCount.Load()
}
