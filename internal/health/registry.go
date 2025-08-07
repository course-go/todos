package health

import (
	"context"
	"fmt"
	"sync"
)

type Registry struct {
	service string
	version string

	mu         sync.Mutex
	components []*Component
}

func NewRegistry(ctx context.Context, opts ...Option) (r *Registry, err error) {
	r = &Registry{}
	for _, opt := range opts {
		err := opt(r)
		if err != nil {
			return nil, fmt.Errorf("failed building registry: %w", err)
		}
	}

	for _, c := range r.components {
		go c.Watch(ctx)
	}

	return r, nil
}

func (r *Registry) RegisterComponent(ctx context.Context, c *Component) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.components = append(r.components, c)
	go c.Watch(ctx)
}

func (r *Registry) Report() Report {
	r.mu.Lock()
	defer r.mu.Unlock()

	health := OK
	cs := make(map[string]ComponentHealth)

	for _, component := range r.components {
		report := component.Report()
		cs[component.name] = report

		if report.Health == ERROR {
			health = ERROR
		}

		if report.Health == WARN && health != ERROR {
			health = WARN
		}
	}

	return Report{
		Service:    r.service,
		Version:    r.version,
		Health:     health,
		Components: cs,
	}
}
