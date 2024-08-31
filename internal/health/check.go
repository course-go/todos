package health

import (
	"context"
	"time"
)

type CheckFn func(context.Context, *Component)

type Check struct {
	Period  time.Duration
	CheckFn CheckFn
}
