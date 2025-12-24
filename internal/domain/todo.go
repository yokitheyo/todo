package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrTodoNotFound       = errors.New("todo not found")
	ErrTitleRequired      = errors.New("title required")
	ErrTitleTooLong       = errors.New("title is too long(max 255)")
	ErrDescriptionTooLong = errors.New("description is too long(max 1_000)")
	ErrInvalidID          = errors.New("invalid id")
	ErrInvalidPath        = errors.New("invalid path")
)

const (
	MaxTitleLength       = 255
	MaxDescriptionLength = 1000
)

type Todo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTodoInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type UpdateTodoInput struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}

type TodoRepository interface {
	Create(ctx context.Context, input CreateTodoInput) (*Todo, error)
	GetByID(ctx context.Context, id int) (*Todo, error)
	GetAll(ctx context.Context) ([]Todo, error)
	Update(ctx context.Context, id int, input UpdateTodoInput) (*Todo, error)
	Delete(ctx context.Context, id int) error
}
