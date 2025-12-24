package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/yokitheyo/todo/internal/domain"
)

type TodoRepository struct {
	mu     sync.RWMutex
	todos  map[int]*domain.Todo
	nextID int
}

func NewTodoRepository() *TodoRepository {
	return &TodoRepository{
		todos:  make(map[int]*domain.Todo),
		nextID: 1,
	}
}

func (r *TodoRepository) Create(ctx context.Context, input domain.CreateTodoInput) (*domain.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	todo := &domain.Todo{
		ID:          r.nextID,
		Title:       input.Title,
		Description: input.Description,
		Completed:   input.Completed,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	r.todos[r.nextID] = todo
	r.nextID++

	return todo, nil
}

// указатель или копию мб датарейс 0_о
func (r *TodoRepository) GetByID(ctx context.Context, id int) (*domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	todo, exists := r.todos[id]
	if !exists {
		return nil, domain.ErrTodoNotFound
	}

	return todo, nil
}

func (r *TodoRepository) GetAll(ctx context.Context) ([]domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	todos := make([]domain.Todo, 0, len(r.todos))
	for _, todo := range r.todos {
		todos = append(todos, *todo)
	}

	return todos, nil
}

func (r *TodoRepository) Update(ctx context.Context, id int, input domain.UpdateTodoInput) (*domain.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	todo, exists := r.todos[id]
	if !exists {
		return nil, domain.ErrTodoNotFound
	}

	if input.Title != nil {
		todo.Title = *input.Title
	}

	if input.Description != nil {
		todo.Description = *input.Description
	}

	if input.Completed != nil {
		todo.Completed = *input.Completed
	}

	todo.UpdatedAt = time.Now()

	return todo, nil
}

func (r *TodoRepository) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.todos[id]; !exists {
		return domain.ErrTodoNotFound
	}

	delete(r.todos, id)
	return nil
}

func (r *TodoRepository) GetFiltered(ctx context.Context, completed *bool, search string) ([]domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Todo
	for _, todo := range r.todos {
		if completed != nil && todo.Completed != *completed {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(todo.Title), strings.ToLower(search)) &&
			!strings.Contains(strings.ToLower(todo.Description), strings.ToLower(search)) {
			continue
		}
		filtered = append(filtered, *todo)
	}

	return filtered, nil
}
