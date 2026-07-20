package ports

import (
	"context"

	domain "go-crud-db-p2/internal/core/domain/platform"
)

type ITodoService interface {
	Create(ctx context.Context, request domain.CreateTodoRequest) (*domain.Todo, error)
	Fetch(ctx context.Context, request domain.FetchTodosRequest) (*domain.TodoList, error)
	GetByID(ctx context.Context, id string, userID domain.UserID) (*domain.Todo, error)
	Update(ctx context.Context, id string, request domain.UpdateTodoRequest, userID domain.UserID) (*domain.Todo, error)
	Delete(ctx context.Context, id string, userID domain.UserID) error
}

type IAuthService interface {
	Register(ctx context.Context, request domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(ctx context.Context, request domain.LoginRequest) (*domain.AuthResponse, error)
	GoogleLoginURL(ctx context.Context) (string, error)
	GoogleCallback(ctx context.Context, state string, code string) (*domain.AuthResponse, error)
	CurrentUser(ctx context.Context, id string) (*domain.User, error)
}
