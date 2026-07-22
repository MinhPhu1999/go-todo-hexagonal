package platform

import (
	"context"
	"errors"
	"time"

	domain "go-crud-db-p2/internal/core/domain/platform"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (repository *UserRepository) Save(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, name, picture, password_hash, google_id, providers, created_at, updated_at, last_login_at, account_locked_until, reset_otp_hash, reset_otp_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			email               = EXCLUDED.email,
			name                = EXCLUDED.name,
			picture             = EXCLUDED.picture,
			password_hash       = EXCLUDED.password_hash,
			google_id           = EXCLUDED.google_id,
			providers           = EXCLUDED.providers,
			updated_at          = EXCLUDED.updated_at,
			last_login_at       = EXCLUDED.last_login_at,
			account_locked_until = EXCLUDED.account_locked_until,
			reset_otp_hash      = EXCLUDED.reset_otp_hash,
			reset_otp_expires_at = EXCLUDED.reset_otp_expires_at
	`

	_, err := repository.pool.Exec(ctx, query,
		user.ID.String(),
		user.Email,
		user.Name,
		user.Picture,
		user.PasswordHash,
		user.GoogleID,
		user.Providers,
		user.CreatedAt,
		user.UpdatedAt,
		user.LastLoginAt,
		user.AccountLockedUntil,
		user.ResetOTPHash,
		user.ResetOTPExpiresAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailAlreadyExists
		}
	}
	return err
}

func (repository *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	query := "SELECT id, email, name, picture, password_hash, google_id, providers, created_at, updated_at, last_login_at, account_locked_until, reset_otp_hash, reset_otp_expires_at FROM users WHERE id = $1"
	return repository.findOne(ctx, query, id.String())
}

func (repository *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	normalized, err := domain.NormalizeEmail(email)
	if err != nil {
		return nil, err
	}
	query := "SELECT id, email, name, picture, password_hash, google_id, providers, created_at, updated_at, last_login_at, account_locked_until, reset_otp_hash, reset_otp_expires_at FROM users WHERE email = $1"
	return repository.findOne(ctx, query, normalized)
}

func (repository *UserRepository) FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	if googleID == "" {
		return nil, domain.ErrUnauthorized
	}
	query := "SELECT id, email, name, picture, password_hash, google_id, providers, created_at, updated_at, last_login_at, account_locked_until, reset_otp_hash, reset_otp_expires_at FROM users WHERE google_id = $1"
	return repository.findOne(ctx, query, googleID)
}

func (repository *UserRepository) findOne(ctx context.Context, query string, args ...interface{}) (*domain.User, error) {
	var document userDocument
	err := scanUserFromDB(repository.pool.QueryRow(ctx, query, args...), &document)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	return document.toDomain()
}

type userDocument struct {
	ID                 string
	Email              string
	Name               string
	Picture            string
	PasswordHash       string
	GoogleID           string
	Providers          []string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	LastLoginAt        *time.Time
	AccountLockedUntil *time.Time
	ResetOTPHash       *string
	ResetOTPExpiresAt  *time.Time
}

func (document userDocument) toDomain() (*domain.User, error) {
	resetOTPHash := ""
	if document.ResetOTPHash != nil {
		resetOTPHash = *document.ResetOTPHash
	}
	return domain.RehydrateUser(
		domain.UserID(document.ID),
		document.Email,
		document.Name,
		document.Picture,
		document.PasswordHash,
		document.GoogleID,
		document.Providers,
		document.CreatedAt,
		document.UpdatedAt,
		document.LastLoginAt,
		document.AccountLockedUntil,
		resetOTPHash,
		document.ResetOTPExpiresAt,
	)
}

func scanUserFromDB(row pgx.Row, document *userDocument) error {
	return row.Scan(
		&document.ID,
		&document.Email,
		&document.Name,
		&document.Picture,
		&document.PasswordHash,
		&document.GoogleID,
		&document.Providers,
		&document.CreatedAt,
		&document.UpdatedAt,
		&document.LastLoginAt,
		&document.AccountLockedUntil,
		&document.ResetOTPHash,
		&document.ResetOTPExpiresAt,
	)
}

func copyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	copied := value.UTC()
	return &copied
}
