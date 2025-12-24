package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yokitheyo/todo/internal/domain"
	"github.com/yokitheyo/todo/pkg/logger"
)

type TodoService interface {
	Create(ctx context.Context, input domain.CreateTodoInput) (*domain.Todo, error)
	GetByID(ctx context.Context, id int) (*domain.Todo, error)
	GetAll(ctx context.Context) ([]domain.Todo, error)
	Update(ctx context.Context, id int, input domain.UpdateTodoInput) (*domain.Todo, error)
	Delete(ctx context.Context, id int) error
	GetFiltered(ctx context.Context, completed *bool, search string) ([]domain.Todo, error)
}

type TodoHandler struct {
	service        TodoService
	log            *logger.Logger
	requestTimeout time.Duration
}

func NewTodoHandler(service TodoService, log *logger.Logger, timeout time.Duration) *TodoHandler {
	return &TodoHandler{
		service:        service,
		log:            log,
		requestTimeout: timeout,
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *TodoHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/todos", h.loggingMiddleware(h.todosHandler))
	mux.HandleFunc("/todos/", h.loggingMiddleware(h.todoByIDHandler))
	mux.HandleFunc("/health", h.healthHandler)
}

func (h *TodoHandler) todosHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()

	switch r.Method {
	case http.MethodPost:
		h.createTodo(ctx, w, r)
	case http.MethodGet:
		if r.URL.Query().Get("completed") != "" || r.URL.Query().Get("search") != "" {
			h.getFilteredTodos(ctx, w, r)
			return
		}
		h.getAllTodos(ctx, w, r)
	default:
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *TodoHandler) todoByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()

	id, err := h.extractID(r.URL.Path)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid todo id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTodoByID(ctx, w, r, id)
	case http.MethodPut:
		h.updateTodo(ctx, w, r, id)
	case http.MethodDelete:
		h.deleteTodo(ctx, w, r, id)
	default:
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *TodoHandler) createTodo(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var input domain.CreateTodoInput
	if err := h.decodeJSON(w, r, &input); err != nil {
		h.handleRequestError(w, err)
		return
	}

	todo, err := h.service.Create(ctx, input)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, todo)
}

func (h *TodoHandler) getAllTodos(ctx context.Context, w http.ResponseWriter, _ *http.Request) {
	todos, err := h.service.GetAll(ctx)
	if err != nil {
		h.log.Error("failed to get todos", "error", err)
		h.respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.respondJSON(w, http.StatusOK, todos)
}

func (h *TodoHandler) getTodoByID(ctx context.Context, w http.ResponseWriter, _ *http.Request, id int) {
	todo, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) updateTodo(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) {
	var input domain.UpdateTodoInput
	if err := h.decodeJSON(w, r, &input); err != nil {
		h.handleRequestError(w, err)
		return
	}

	todo, err := h.service.Update(ctx, id, input)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) deleteTodo(ctx context.Context, w http.ResponseWriter, _ *http.Request, id int) {
	err := h.service.Delete(ctx, id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TodoHandler) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *TodoHandler) handleRequestError(w http.ResponseWriter, err error) {
	var syntaxErr *json.SyntaxError
	var unmarshalTypeErr *json.UnmarshalTypeError

	switch {
	case errors.As(err, &syntaxErr):
		h.respondError(w, http.StatusBadRequest, "malformed JSON at position "+strconv.Itoa(int(syntaxErr.Offset)))
	case errors.As(err, &unmarshalTypeErr):
		h.respondError(w, http.StatusBadRequest, "invalid value for field "+unmarshalTypeErr.Field)
	case errors.Is(err, io.EOF):
		h.respondError(w, http.StatusBadRequest, "empty request body")
	default:
		h.respondError(w, http.StatusBadRequest, "invalid request body")
	}
}

func (h *TodoHandler) getFilteredTodos(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	completedStr := query.Get("completed")
	search := query.Get("search")

	var completed *bool
	if completedStr != "" {
		b := completedStr == "true"
		completed = &b
	}

	todos, err := h.service.GetFiltered(ctx, completed, search)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, todos)
}

func (h *TodoHandler) decodeJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func (h *TodoHandler) extractID(path string) (int, error) {
	path = strings.TrimPrefix(path, "/todos/")
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		return 0, domain.ErrInvalidPath
	}

	id, err := strconv.Atoi(path)
	if err != nil || id <= 0 {
		return 0, domain.ErrInvalidID
	}

	return id, nil
}

func (h *TodoHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrTodoNotFound):
		h.respondError(w, http.StatusNotFound, "todo not found")
	case errors.Is(err, domain.ErrTitleRequired),
		errors.Is(err, domain.ErrTitleTooLong),
		errors.Is(err, domain.ErrDescriptionTooLong),
		errors.Is(err, domain.ErrInvalidID):
		h.respondError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("service error", "error", err, "operation", "unknown")
		h.respondError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *TodoHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.log.Error("failed to encode response", "error", err)
		}
	}
}

func (h *TodoHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, errorResponse{Error: message})
}

func (h *TodoHandler) loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		userAgent := r.UserAgent()
		sw := &statusResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		h.log.Info("incoming request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"user_agent", userAgent,
		)

		next(sw, r)

		h.log.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", time.Since(start),
		)
	}
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
