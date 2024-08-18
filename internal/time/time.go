package time

import "time"

// Factory is a type alias for time factory.
// It is used for stubbing the time "Now" implementation.
type Factory func() time.Time
