package health

import (
	"context"
	"sync"
	"time"
)

type Component struct {
	name   string
	checks []Check

	mu        sync.Mutex
	Health    Health
	Message   string
	UpdatedAt time.Time
}

func NewComponent(name string, checks ...Check) *Component {
	return &Component{
		name:   name,
		checks: checks,
	}
}

func (c *Component) Watch(ctx context.Context) {
	for _, check := range c.checks {
		go func() {
			t := time.NewTicker(check.Period)
			for {
				check.CheckFn(ctx, c)

				select {
				case <-t.C:

				case <-ctx.Done():
				}
			}
		}()
	}
}

func (c *Component) Report() ComponentHealth {
	c.mu.Lock()
	defer c.mu.Unlock()
	return ComponentHealth{
		Health:    c.Health,
		Message:   c.Message,
		UpdatedAt: c.UpdatedAt,
	}
}

type ComponentHealth struct {
	Health    Health    `json:"health"`
	Message   string    `json:"message,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}
