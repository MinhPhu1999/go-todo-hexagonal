package platform

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "go-crud-db-p2/internal/core/domain/platform"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoRepository struct {
	pool *pgxpool.Pool
}

func NewTodoRepository(pool *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{pool: pool}
}

func (repository *TodoRepository) Save(ctx context.Context, todo *domain.Todo) error {
	query := `
		INSERT INTO todos (id, user_id, title, description, completed, created_at, updated_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			user_id      = EXCLUDED.user_id,
			title        = EXCLUDED.title,
			description  = EXCLUDED.description,
			completed    = EXCLUDED.completed,
			updated_at   = EXCLUDED.updated_at,
			completed_at = EXCLUDED.completed_at
	`

	_, err := repository.pool.Exec(ctx, query,
		todo.ID.String(),
		todo.UserID.String(),
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.CreatedAt,
		todo.UpdatedAt,
		todo.CompletedAt,
	)
	return err
}

func (repository *TodoRepository) Fetch(ctx context.Context, request domain.FetchTodosRequest) (*domain.TodoList, error) {
	args := make([]interface{}, 0)
	argIdx := 1

	where := fmt.Sprintf("WHERE user_id = $%d", argIdx)
	args = append(args, request.UserID.String())
	argIdx++

	if request.Completed != nil {
		where = fmt.Sprintf("%s AND completed = $%d", where, argIdx)
		args = append(args, *request.Completed)
		argIdx++
	}

	if request.Search != "" {
		searchClause := fmt.Sprintf(
			"AND (title ILIKE $%d OR description ILIKE $%d)",
			argIdx, argIdx+1,
		)
		where = fmt.Sprintf("%s %s", where, searchClause)
		searchPattern := "%" + request.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIdx += 2
	}

	countQuery := "SELECT COUNT(*) FROM todos " + where
	var total int64
	if err := repository.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (request.PageNumber - 1) * request.PageSize
	dataQuery := fmt.Sprintf(
		"SELECT id, user_id, title, description, completed, created_at, updated_at, completed_at FROM todos %s ORDER BY created_at DESC OFFSET $%d LIMIT $%d",
		where, argIdx, argIdx+1,
	)
	args = append(args, offset, request.PageSize)

	rows, err := repository.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := make([]*domain.Todo, 0)
	for rows.Next() {
		document, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}

		todo, err := document.toDomain()
		if err != nil {
			return nil, err
		}

		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.TodoList{
		Items:      todos,
		PageSize:   request.PageSize,
		PageNumber: request.PageNumber,
		Total:      total,
	}, nil
}

func (repository *TodoRepository) GetByID(ctx context.Context, id domain.TodoID, userID domain.UserID) (*domain.Todo, error) {
	query := "SELECT id, user_id, title, description, completed, created_at, updated_at, completed_at FROM todos WHERE id = $1 AND user_id = $2"

	var document todoDocument
	err := scanTodoFromDB(repository.pool.QueryRow(ctx, query, id.String(), userID.String()), &document)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTodoNotFound
	}
	if err != nil {
		return nil, err
	}

	return document.toDomain()
}

func (repository *TodoRepository) Delete(ctx context.Context, id domain.TodoID, userID domain.UserID) error {
	query := "DELETE FROM todos WHERE id = $1 AND user_id = $2"

	result, err := repository.pool.Exec(ctx, query, id.String(), userID.String())
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrTodoNotFound
	}

	return nil
}

type todoDocument struct {
	ID          string
	UserID      string
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

func (document todoDocument) toDomain() (*domain.Todo, error) {
	return domain.RehydrateTodo(
		domain.TodoID(document.ID),
		domain.UserID(document.UserID),
		document.Title,
		document.Description,
		document.Completed,
		document.CreatedAt,
		document.UpdatedAt,
		document.CompletedAt,
	)
}

type scannableTodo interface {
	Scan(dest ...interface{}) error
}

func scanTodo(row scannableTodo) (todoDocument, error) {
	var document todoDocument
	err := row.Scan(
		&document.ID,
		&document.UserID,
		&document.Title,
		&document.Description,
		&document.Completed,
		&document.CreatedAt,
		&document.UpdatedAt,
		&document.CompletedAt,
	)
	return document, err
}

func scanTodoFromDB(row pgx.Row, document *todoDocument) error {
	return row.Scan(
		&document.ID,
		&document.UserID,
		&document.Title,
		&document.Description,
		&document.Completed,
		&document.CreatedAt,
		&document.UpdatedAt,
		&document.CompletedAt,
	)
}
