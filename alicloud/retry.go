package alicloud

import (
	"time"
)

// ExpBackoffWait returns a closure that sleeps with exponential backoff and jitter.
// initial: starting delay; factor: growth multiplier; jitterRatio: +/- percentage (e.g., 0.25 for 25%);
// maxDelay: upper bound for sleep per attempt.
// Note: This helper does not manage total timeout; callers should wrap it with resource.Retry respecting operation timeouts.
func ExpBackoffWait(initial time.Duration, factor float64, jitterRatio float64, maxDelay time.Duration) func() {
	attempt := 0
	return func() {
		// Compute delay = initial * factor^attempt (capped to maxDelay)
		delay := float64(initial)
		if attempt > 0 {
			delay = float64(initial)
			for i := 0; i < attempt; i++ {
				delay *= factor
			}
		}
		if time.Duration(delay) > maxDelay {
			delay = float64(maxDelay)
		}
		// Apply jitter in range [1-jitterRatio, 1+jitterRatio]
		if jitterRatio > 0 {
			nsec := time.Now().UnixNano()
			sign := 1.0
			if nsec&1 == 1 {
				sign = -1.0
			}
			frac := float64((nsec%997)+1) / 997.0 // (0,1]
			jitter := sign * jitterRatio * frac
			delay = delay * (1.0 + jitter)
			if delay < float64(initial) {
				delay = float64(initial)
			}
		}
		time.Sleep(time.Duration(delay))
		attempt++
	}
}

// IsOssConcurrentUpdateError returns true if the error indicates an OSS concurrent update (HTTP 409)
// Along with general retryable errors (NeedRetry), this can be used to gate retries in write operations.
func IsOssConcurrentUpdateError(err error) bool {
	if err == nil {
		return false
	}
	if IsExpectedErrors(err, []string{"ConcurrentUpdateBucketFailed"}) {
		return true
	}
	return NeedRetry(err)
}
