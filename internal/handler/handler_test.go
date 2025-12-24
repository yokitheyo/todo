package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/yokitheyo/todo/internal/domain"
	"github.com/yokitheyo/todo/internal/repository/memory"
	"github.com/yokitheyo/todo/internal/service"
	"github.com/yokitheyo/todo/pkg/logger"
)

func setupTestHandler(t *testing.T) (*TodoHandler, *memory.TodoRepository) {
	repo := memory.NewTodoRepository()
	svc := service.NewTodoService(repo)
	log := logger.New("error", nil, "json")
	handler := NewTodoHandler(svc, log, 2*time.Second)
	return handler, repo
}

type TodoServiceMock struct {
	Repo *memory.TodoRepository
}

func (s *TodoServiceMock) Create(ctx context.Context, input domain.CreateTodoInput) (*domain.Todo, error) {
	return s.Repo.Create(ctx, input)
}
func (s *TodoServiceMock) GetByID(ctx context.Context, id int) (*domain.Todo, error) {
	return s.Repo.GetByID(ctx, id)
}
func (s *TodoServiceMock) GetAll(ctx context.Context) ([]domain.Todo, error) {
	return s.Repo.GetAll(ctx)
}
func (s *TodoServiceMock) Update(ctx context.Context, id int, input domain.UpdateTodoInput) (*domain.Todo, error) {
	return s.Repo.Update(ctx, id, input)
}
func (s *TodoServiceMock) Delete(ctx context.Context, id int) error {
	return s.Repo.Delete(ctx, id)
}

func TestTodoHandler_CreateGetUpdateDelete(t *testing.T) {
	handler, _ := setupTestHandler(t)

	// post
	input := map[string]interface{}{
		"title":       "Task 1",
		"description": "Desc 1",
		"completed":   false,
	}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.todosHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", w.Code)
	}

	var todo domain.Todo
	if err := json.NewDecoder(w.Body).Decode(&todo); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	// get
	req = httptest.NewRequest(http.MethodGet, "/todos/"+strconv.Itoa(todo.ID), nil)
	w = httptest.NewRecorder()
	handler.todoByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var fetched domain.Todo
	if err := json.NewDecoder(w.Body).Decode(&fetched); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if fetched.ID != todo.ID {
		t.Errorf("expected ID=%d, got %d", todo.ID, fetched.ID)
	}

	// put
	updateInput := map[string]interface{}{
		"title":       "Updated Task",
		"description": "Updated Desc",
		"completed":   true,
	}
	body, _ = json.Marshal(updateInput)
	req = httptest.NewRequest(http.MethodPut, "/todos/"+strconv.Itoa(todo.ID), bytes.NewReader(body))
	w = httptest.NewRecorder()
	handler.todoByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var updated domain.Todo
	if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if updated.Title != "Updated Task" || !updated.Completed {
		t.Errorf("update not applied correctly")
	}

	// del
	req = httptest.NewRequest(http.MethodDelete, "/todos/"+strconv.Itoa(todo.ID), nil)
	w = httptest.NewRecorder()
	handler.todoByIDHandler(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", w.Code)
	}

	// get
	req = httptest.NewRequest(http.MethodGet, "/todos/"+strconv.Itoa(todo.ID), nil)
	w = httptest.NewRecorder()
	handler.todoByIDHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", w.Code)
	}
}

func TestTodoHandler_ValidationErrors(t *testing.T) {
	handler, _ := setupTestHandler(t)

	input := map[string]interface{}{
		"title":       "",
		"description": "Desc",
	}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.todosHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for empty title, got %d", w.Code)
	}
}

func TestTodoHandler_NotFound(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/todos/999", nil)
	w := httptest.NewRecorder()
	handler.todoByIDHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", w.Code)
	}
}

func TestTodoHandler_GetAllEmpty(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	w := httptest.NewRecorder()
	handler.todosHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var todos []domain.Todo
	if err := json.NewDecoder(w.Body).Decode(&todos); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(todos) != 0 {
		t.Errorf("expected 0 todos, got %d", len(todos))
	}
}
