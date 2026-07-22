package ports

import (
	"context"
	"time"

	domain "go-crud-db-p2/internal/core/domain/platform"
)

type ITodoRepository interface {
	Save(ctx context.Context, todo *domain.Todo) error
	Fetch(ctx context.Context, request domain.FetchTodosRequest) (*domain.TodoList, error)
	GetByID(ctx context.Context, id domain.TodoID, userID domain.UserID) (*domain.Todo, error)
	Delete(ctx context.Context, id domain.TodoID, userID domain.UserID) error
}

type IUserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error)
}

type ITodoIDGenerator interface {
	NewID() domain.TodoID
}

type IUserIDGenerator interface {
	NewUserID() domain.UserID
}

type IPasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) error
}

type ITokenProvider interface {
	IssueToken(user *domain.User) (domain.AuthToken, error)
	VerifyToken(token string) (domain.TokenClaims, error)
}

type IGoogleIdentityProvider interface {
	AuthCodeURL(state string) (string, error)
	Exchange(ctx context.Context, code string) (domain.GoogleProfile, error)
}

type IAuthStateStore interface {
	Generate() (string, error)
	Verify(state string) bool
}

type IClock interface {
	Now() time.Time
}

type IEmailSender interface {
	SendOTP(to string, otp string) error
}
