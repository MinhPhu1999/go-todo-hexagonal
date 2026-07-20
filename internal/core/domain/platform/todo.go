package platform

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	maxTitleLength       = 160
	maxDescriptionLength = 2000
	defaultTodoPageSize  = 10
	defaultTodoPage      = 1
	maxTodoPageSize      = 100
)

var (
	ErrInvalidTodo  = errors.New("invalid todo")
	ErrTodoNotFound = errors.New("todo not found")
)

type TodoID string

func ParseTodoID(value string) (TodoID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%w: id is required", ErrInvalidTodo)
	}
	return TodoID(value), nil
}

func (id TodoID) String() string {
	return string(id)
}

type Todo struct {
	ID          TodoID     `json:"id"`
	UserID      UserID     `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

func NewTodo(id TodoID, userID UserID, title string, description string, now time.Time) (*Todo, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidTodo)
	}
	if userID == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidTodo)
	}

	title, err := normalizeTitle(title)
	if err != nil {
		return nil, err
	}

	description, err = normalizeDescription(description)
	if err != nil {
		return nil, err
	}

	now = now.UTC()
	return &Todo{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func RehydrateTodo(
	id TodoID,
	userID UserID,
	title string,
	description string,
	completed bool,
	createdAt time.Time,
	updatedAt time.Time,
	completedAt *time.Time,
) (*Todo, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidTodo)
	}

	title, err := normalizeTitle(title)
	if err != nil {
		return nil, err
	}

	description, err = normalizeDescription(description)
	if err != nil {
		return nil, err
	}

	if !completed {
		completedAt = nil
	}

	return &Todo{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Description: description,
		Completed:   completed,
		CreatedAt:   createdAt.UTC(),
		UpdatedAt:   updatedAt.UTC(),
		CompletedAt: copyTimePtr(completedAt),
	}, nil
}

func (todo *Todo) Update(title *string, description *string, completed *bool, now time.Time) error {
	changed := false

	if title != nil {
		value, err := normalizeTitle(*title)
		if err != nil {
			return err
		}
		if todo.Title != value {
			todo.Title = value
			changed = true
		}
	}

	if description != nil {
		value, err := normalizeDescription(*description)
		if err != nil {
			return err
		}
		if todo.Description != value {
			todo.Description = value
			changed = true
		}
	}

	if completed != nil {
		if *completed && !todo.Completed {
			completedAt := now.UTC()
			todo.Completed = true
			todo.CompletedAt = &completedAt
			changed = true
		}

		if !*completed && todo.Completed {
			todo.Completed = false
			todo.CompletedAt = nil
			changed = true
		}
	}

	if changed {
		todo.UpdatedAt = now.UTC()
	}

	return nil
}

type CreateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	UserID      UserID `json:"-"`
}

type UpdateTodoRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}

type FetchTodosRequest struct {
	PageSize   int
	PageNumber int
	Completed  *bool
	Search     string
	UserID     UserID
}

func (request FetchTodosRequest) Normalize() (FetchTodosRequest, error) {
	if request.PageSize == 0 {
		request.PageSize = defaultTodoPageSize
	}
	if request.PageNumber == 0 {
		request.PageNumber = defaultTodoPage
	}
	request.Search = strings.TrimSpace(request.Search)

	if request.PageSize < 1 {
		return FetchTodosRequest{}, fmt.Errorf("%w: pageSize must be greater than 0", ErrInvalidTodo)
	}
	if request.PageSize > maxTodoPageSize {
		return FetchTodosRequest{}, fmt.Errorf("%w: pageSize must be at most %d", ErrInvalidTodo, maxTodoPageSize)
	}
	if request.PageNumber < 1 {
		return FetchTodosRequest{}, fmt.Errorf("%w: pageNumber must be greater than 0", ErrInvalidTodo)
	}

	return request, nil
}

type TodoList struct {
	Items      []*Todo
	PageSize   int
	PageNumber int
	Total      int64
}

func (list TodoList) TotalPages() int {
	if list.Total == 0 || list.PageSize == 0 {
		return 0
	}

	pageSize := int64(list.PageSize)
	return int((list.Total + pageSize - 1) / pageSize)
}

func normalizeTitle(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%w: title is required", ErrInvalidTodo)
	}
	if len(value) > maxTitleLength {
		return "", fmt.Errorf("%w: title must be at most %d characters", ErrInvalidTodo, maxTitleLength)
	}
	return value, nil
}

func normalizeDescription(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) > maxDescriptionLength {
		return "", fmt.Errorf("%w: description must be at most %d characters", ErrInvalidTodo, maxDescriptionLength)
	}
	return value, nil
}

func copyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	copied := value.UTC()
	return &copied
}
