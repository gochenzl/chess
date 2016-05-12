package util

import "time"

func ClockMonotonic() time.Time {
	return time.Now()
}
