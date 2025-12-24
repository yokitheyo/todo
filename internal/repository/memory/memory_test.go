package memory

import (
	"context"
	"strconv"
	"testing"

	"github.com/yokitheyo/todo/internal/domain"
)

func TestTodoRepository_CreateGetUpdateDelete(t *testing.T) {
	repo := NewTodoRepository()
	ctx := context.Background()

	input := domain.CreateTodoInput{
		Title:       "Test Task",
		Description: "Test Description",
		Completed:   false,
	}

	todo, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if todo.ID != 1 {
		t.Errorf("expected ID=1, got %d", todo.ID)
	}
	if todo.Title != input.Title {
		t.Errorf("expected Title=%s, got %s", input.Title, todo.Title)
	}

	got, err := repo.GetByID(ctx, todo.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.ID != todo.ID {
		t.Errorf("expected ID=%d, got %d", todo.ID, got.ID)
	}

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 todo, got %d", len(all))
	}

	newTitle := "Updated Title"
	newDesc := "Updated Desc"
	newCompleted := true
	updateInput := domain.UpdateTodoInput{
		Title:       &newTitle,
		Description: &newDesc,
		Completed:   &newCompleted,
	}

	updated, err := repo.Update(ctx, todo.ID, updateInput)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Title != newTitle || updated.Description != newDesc || updated.Completed != newCompleted {
		t.Errorf("update did not apply correctly")
	}
	if updated.UpdatedAt.Before(updated.CreatedAt) {
		t.Errorf("UpdatedAt not updated correctly")
	}

	if err := repo.Delete(ctx, todo.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, todo.ID)
	if err != domain.ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound after delete, got %v", err)
	}

	if err := repo.Delete(ctx, todo.ID); err != domain.ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound when deleting non-existing todo, got %v", err)
	}
}

func TestTodoRepository_Concurrency(t *testing.T) {
	repo := NewTodoRepository()
	ctx := context.Background()

	const n = 100
	done := make(chan struct{})

	for i := 0; i < n; i++ {
		go func(i int) {
			input := domain.CreateTodoInput{
				Title:       "Task " + strconv.Itoa(i),
				Description: "Desc",
				Completed:   false,
			}
			_, _ = repo.Create(ctx, input)
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < n; i++ {
		<-done
	}

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(all) != n {
		t.Errorf("expected %d todos, got %d", n, len(all))
	}
}
