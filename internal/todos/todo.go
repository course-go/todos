package todos

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID          uuid.UUID  `json:"id,omitempty"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"createdAt,omitzero"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}
