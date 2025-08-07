package request

import "time"

type CreateTodoRequest struct {
	Description string `json:"description" validate:"required"`
}

type UpdateTodoRequest struct {
	Description string     `json:"description" validate:"required"`
	CompletedAt *time.Time `json:"completedAt"`
}
