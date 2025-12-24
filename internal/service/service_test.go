package service_test

import (
	"context"
	"testing"

	"github.com/yokitheyo/todo/internal/domain"
	"github.com/yokitheyo/todo/internal/repository/memory"
	"github.com/yokitheyo/todo/internal/service"
)

func setupService() (*service.TodoService, *memory.TodoRepository) {
	repo := memory.NewTodoRepository()
	svc := service.NewTodoService(repo)
	return svc, repo
}

func TestCreateTodo_Success(t *testing.T) {
	svc, _ := setupService()
	input := domain.CreateTodoInput{
		Title:       "Test todo",
		Description: "Test description",
		Completed:   false,
	}

	todo, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if todo.ID == 0 {
		t.Fatalf("expected ID to be set, got %d", todo.ID)
	}
	if todo.Title != input.Title {
		t.Errorf("expected title %q, got %q", input.Title, todo.Title)
	}
}

func TestCreateTodo_EmptyTitle(t *testing.T) {
	svc, _ := setupService()
	input := domain.CreateTodoInput{
		Title:       "   ",
		Description: "desc",
	}

	_, err := svc.Create(context.Background(), input)
	if err != domain.ErrTitleRequired {
		t.Errorf("expected ErrTitleRequired, got %v", err)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc, _ := setupService()

	_, err := svc.GetByID(context.Background(), 999)
	if err != domain.ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound, got %v", err)
	}
}

func TestUpdateTodo_Success(t *testing.T) {
	svc, repo := setupService()
	todo, _ := svc.Create(context.Background(), domain.CreateTodoInput{
		Title:       "Original",
		Description: "desc",
	})

	newTitle := "Updated"
	newDesc := "New description"
	updated, err := svc.Update(context.Background(), todo.ID, domain.UpdateTodoInput{
		Title:       &newTitle,
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Title != newTitle {
		t.Errorf("expected title %q, got %q", newTitle, updated.Title)
	}
	if updated.Description != newDesc {
		t.Errorf("expected desc %q, got %q", newDesc, updated.Description)
	}

	got, _ := repo.GetByID(context.Background(), todo.ID)
	if got.Title != newTitle {
		t.Errorf("repo not updated, expected %q got %q", newTitle, got.Title)
	}
}

func TestDeleteTodo(t *testing.T) {
	svc, _ := setupService()
	todo, _ := svc.Create(context.Background(), domain.CreateTodoInput{
		Title: "To delete",
	})

	err := svc.Delete(context.Background(), todo.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = svc.GetByID(context.Background(), todo.ID)
	if err != domain.ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound after delete, got %v", err)
	}
}

func TestUpdateTodo_InvalidID(t *testing.T) {
	svc, _ := setupService()
	newTitle := "Updated"
	_, err := svc.Update(context.Background(), -1, domain.UpdateTodoInput{
		Title: &newTitle,
	})
	if err != domain.ErrInvalidID {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}
}

func TestCreateTodo_TitleTooLong(t *testing.T) {
	svc, _ := setupService()
	longTitle := ""
	for i := 0; i < domain.MaxTitleLength+1; i++ {
		longTitle += "a"
	}
	_, err := svc.Create(context.Background(), domain.CreateTodoInput{
		Title: longTitle,
	})
	if err != domain.ErrTitleTooLong {
		t.Errorf("expected ErrTitleTooLong, got %v", err)
	}
}
