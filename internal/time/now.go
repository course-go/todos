package time

import "time"

func Now() Factory {
	return time.Now
}
