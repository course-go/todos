package time

import "time"

func Now() Factory {
	return func() time.Time {
		return time.Now()
	}
}
