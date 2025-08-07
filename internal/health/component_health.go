package health

import "time"

type ComponentHealth struct {
	Health    Health    `json:"health"`
	Message   string    `json:"message,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}
