package service

import (
	"context"
	"strings"

	"github.com/yokitheyo/todo/internal/domain"
)

type TodoService struct {
	repo domain.TodoRepository
}

func NewTodoService(repo domain.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

func (s *TodoService) Create(ctx context.Context, input domain.CreateTodoInput) (*domain.Todo, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if err := s.validateCreateInput(input); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, input)
}

func (s *TodoService) GetByID(ctx context.Context, id int) (*domain.Todo, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *TodoService) GetAll(ctx context.Context) ([]domain.Todo, error) {
	return s.repo.GetAll(ctx)
}

func (s *TodoService) Update(ctx context.Context, id int, input domain.UpdateTodoInput) (*domain.Todo, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}

	if err := s.validateUpdateInput(input); err != nil {
		return nil, err
	}

	if input.Title != nil {
		trimmed := strings.TrimSpace(*input.Title)
		input.Title = &trimmed
	}

	if input.Description != nil {
		trimmed := strings.TrimSpace(*input.Description)
		input.Description = &trimmed
	}

	return s.repo.Update(ctx, id, input)
}

func (s *TodoService) Delete(ctx context.Context, id int) error {
	if err := validateID(id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *TodoService) validateTitle(title string) error {
	t := strings.TrimSpace(title)
	if t == "" {
		return domain.ErrTitleRequired
	}
	if len(t) > domain.MaxTitleLength {
		return domain.ErrTitleTooLong
	}
	return nil
}

func (s *TodoService) validateDescription(desc string) error {
	if len(desc) > domain.MaxDescriptionLength {
		return domain.ErrDescriptionTooLong
	}
	return nil
}

func (s *TodoService) validateCreateInput(input domain.CreateTodoInput) error {
	if err := s.validateTitle(input.Title); err != nil {
		return err
	}

	if err := s.validateDescription(input.Description); err != nil {
		return err
	}

	return nil
}

func (s *TodoService) validateUpdateInput(input domain.UpdateTodoInput) error {
	if input.Title != nil && s.validateTitle(*input.Title) != nil {
		return s.validateTitle(*input.Title)
	}

	if input.Description != nil {
		if err := s.validateDescription(*input.Description); err != nil {
			return err
		}
	}

	return nil
}

func validateID(id int) error {
	if id <= 0 {
		return domain.ErrInvalidID
	}
	return nil
}
