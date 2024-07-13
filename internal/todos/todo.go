package todos

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID          uuid.UUID  `json:"id,omitempty"`
	Description string     `json:"description,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}
