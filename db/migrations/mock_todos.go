package migrations

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

//go:embed mockdata/todos.json
var mockTodosFS embed.FS

type mockTodo struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

func insertMockTodos(ctx context.Context, database *mongo.Database, config Config) error {
	todos, err := loadMockTodos()
	if err != nil {
		return err
	}

	collection := database.Collection(config.TodoCollection)
	for _, todo := range todos {
		document, err := todo.toDocument()
		if err != nil {
			return err
		}

		_, err = collection.ReplaceOne(
			ctx,
			bson.M{"_id": document["_id"]},
			document,
			options.Replace().SetUpsert(true),
		)
		if err != nil {
			return fmt.Errorf("upsert mock todo %s: %w", todo.ID, err)
		}
	}

	return nil
}

func deleteMockTodos(ctx context.Context, database *mongo.Database, config Config) error {
	todos, err := loadMockTodos()
	if err != nil {
		return err
	}

	ids := make([]bson.ObjectID, 0, len(todos))
	for _, todo := range todos {
		objectID, err := bson.ObjectIDFromHex(todo.ID)
		if err != nil {
			return fmt.Errorf("parse mock todo id %s: %w", todo.ID, err)
		}
		ids = append(ids, objectID)
	}

	_, err = database.Collection(config.TodoCollection).DeleteMany(ctx, bson.M{
		"_id": bson.M{"$in": ids},
	})
	return err
}

func loadMockTodos() ([]mockTodo, error) {
	content, err := mockTodosFS.ReadFile("mockdata/todos.json")
	if err != nil {
		return nil, fmt.Errorf("read mock todos: %w", err)
	}

	var todos []mockTodo
	if err := json.Unmarshal(content, &todos); err != nil {
		return nil, fmt.Errorf("decode mock todos: %w", err)
	}

	return todos, nil
}

func (todo mockTodo) toDocument() (bson.M, error) {
	objectID, err := bson.ObjectIDFromHex(todo.ID)
	if err != nil {
		return nil, fmt.Errorf("parse mock todo id %s: %w", todo.ID, err)
	}

	return bson.M{
		"_id":          objectID,
		"title":        todo.Title,
		"description":  todo.Description,
		"completed":    todo.Completed,
		"created_at":   todo.CreatedAt.UTC(),
		"updated_at":   todo.UpdatedAt.UTC(),
		"completed_at": copyTimePtr(todo.CompletedAt),
	}, nil
}

func copyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	copied := value.UTC()
	return &copied
}
