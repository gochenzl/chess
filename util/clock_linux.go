package util

import "time"
import "github.com/davecheney/junk/clock"

func ClockMonotonic() time.Time {
	return clock.Monotonic.Now()
}
