//go:build expbackoff_unit
// +build expbackoff_unit

package alicloud

import (
	"testing"
	"time"
)

// TestExpBackoffWaitBasic verifies the exponential growth and capping behavior roughly.
func TestExpBackoffWaitBasic(t *testing.T) {
	initial := 2 * time.Millisecond
	factor := 2.0
	jitter := 0.0 // disable jitter for deterministic timing
	maxDelay := 8 * time.Millisecond

	wait := ExpBackoffWait(initial, factor, jitter, maxDelay)

	start := time.Now()
	// attempt 0 -> 2ms
	wait()
	// attempt 1 -> 4ms
	wait()
	// attempt 2 -> 8ms (capped at maxDelay)
	wait()
	// total expected ~14ms; allow generous tolerance for scheduling
	elapsed := time.Since(start)
	if elapsed < 10*time.Millisecond {
		t.Fatalf("elapsed too short: %v", elapsed)
	}
	if elapsed > 50*time.Millisecond {
		t.Fatalf("elapsed too long: %v", elapsed)
	}
}

func TestIsOssConcurrentUpdateErrorNil(t *testing.T) {
	if IsOssConcurrentUpdateError(nil) {
		t.Fatalf("nil error should not be considered concurrent update error")
	}
}
