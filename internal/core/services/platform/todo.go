package platform

import (
	"context"
	"fmt"
	"time"

	"go-crud-db-p2/config"
	domain "go-crud-db-p2/internal/core/domain/platform"
	ports "go-crud-db-p2/internal/core/ports/platform"
)

type TodoService struct {
	todoRepository ports.ITodoRepository
	contextTimeout time.Duration
	idGenerator    ports.ITodoIDGenerator
	clock          ports.IClock
}

func NewTodoService(
	todoRepository ports.ITodoRepository,
	cfg config.Config,
	idGenerator ports.ITodoIDGenerator,
	clock ports.IClock,
) *TodoService {
	return &TodoService{
		todoRepository: todoRepository,
		contextTimeout: cfg.Context.Timeout,
		idGenerator:    idGenerator,
		clock:          clock,
	}
}

func (s *TodoService) Create(ctx context.Context, request domain.CreateTodoRequest) (*domain.Todo, error) {
	ctx, cancel := s.context(ctx)
	defer cancel()

	todo, err := domain.NewTodo(s.idGenerator.NewID(), request.UserID, request.Title, request.Description, s.clock.Now())
	if err != nil {
		return nil, err
	}

	if err := s.todoRepository.Save(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

func (s *TodoService) Fetch(ctx context.Context, request domain.FetchTodosRequest) (*domain.TodoList, error) {
	ctx, cancel := s.context(ctx)
	defer cancel()

	request, err := request.Normalize()
	if err != nil {
		return nil, err
	}

	return s.todoRepository.Fetch(ctx, request)
}

func (s *TodoService) GetByID(ctx context.Context, id string, userID domain.UserID) (*domain.Todo, error) {
	ctx, cancel := s.context(ctx)
	defer cancel()

	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, err
	}

	return s.todoRepository.GetByID(ctx, todoID, userID)
}

func (s *TodoService) Update(ctx context.Context, id string, request domain.UpdateTodoRequest, userID domain.UserID) (*domain.Todo, error) {
	if request.Title == nil && request.Description == nil && request.Completed == nil {
		return nil, fmt.Errorf("%w: at least one field is required", domain.ErrInvalidTodo)
	}

	ctx, cancel := s.context(ctx)
	defer cancel()

	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, err
	}

	todo, err := s.todoRepository.GetByID(ctx, todoID, userID)
	if err != nil {
		return nil, err
	}

	if err := todo.Update(request.Title, request.Description, request.Completed, s.clock.Now()); err != nil {
		return nil, err
	}

	if err := s.todoRepository.Save(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

func (s *TodoService) Delete(ctx context.Context, id string, userID domain.UserID) error {
	ctx, cancel := s.context(ctx)
	defer cancel()

	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return err
	}

	return s.todoRepository.Delete(ctx, todoID, userID)
}

func (s *TodoService) context(parent context.Context) (context.Context, context.CancelFunc) {
	if s.contextTimeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, s.contextTimeout)
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}
