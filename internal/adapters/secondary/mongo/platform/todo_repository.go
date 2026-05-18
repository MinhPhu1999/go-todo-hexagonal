package platform

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"go-crud-db-p2/config"
	domain "go-crud-db-p2/internal/core/domain/platform"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type TodoRepository struct {
	collection *mongo.Collection
}

func NewTodoRepository(database *mongo.Database, collection config.TodoCollectionName) *TodoRepository {
	return &TodoRepository{
		collection: database.Collection(string(collection)),
	}
}

func (repository *TodoRepository) Save(ctx context.Context, todo *domain.Todo) error {
	document, err := newTodoDocument(todo)
	if err != nil {
		return err
	}

	_, err = repository.collection.ReplaceOne(
		ctx,
		bson.M{"_id": document.ID},
		document,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (repository *TodoRepository) Fetch(ctx context.Context, request domain.FetchTodosRequest) (*domain.TodoList, error) {
	filter := bson.D{}
	if request.Completed != nil {
		filter = append(filter, bson.E{Key: "completed", Value: *request.Completed})
	}
	if request.Search != "" {
		search := bson.Regex{Pattern: regexp.QuoteMeta(request.Search), Options: "i"}
		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "title", Value: search}},
			bson.D{{Key: "description", Value: search}},
		}})
	}

	total, err := repository.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	skip := int64(request.PageNumber-1) * int64(request.PageSize)
	cursor, err := repository.collection.Find(
		ctx,
		filter,
		options.Find().
			SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(skip).
			SetLimit(int64(request.PageSize)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	todos := make([]*domain.Todo, 0)
	for cursor.Next(ctx) {
		var document todoDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}

		todo, err := document.toDomain()
		if err != nil {
			return nil, err
		}

		todos = append(todos, todo)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return &domain.TodoList{
		Items:      todos,
		PageSize:   request.PageSize,
		PageNumber: request.PageNumber,
		Total:      total,
	}, nil
}

func (repository *TodoRepository) GetByID(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
	objectID, err := objectIDFromTodoID(id)
	if err != nil {
		return nil, err
	}

	var document todoDocument
	err = repository.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrTodoNotFound
	}
	if err != nil {
		return nil, err
	}

	return document.toDomain()
}

func (repository *TodoRepository) Delete(ctx context.Context, id domain.TodoID) error {
	objectID, err := objectIDFromTodoID(id)
	if err != nil {
		return err
	}

	result, err := repository.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrTodoNotFound
	}

	return nil
}

type todoDocument struct {
	ID          bson.ObjectID `bson:"_id"`
	Title       string        `bson:"title"`
	Description string        `bson:"description"`
	Completed   bool          `bson:"completed"`
	CreatedAt   time.Time     `bson:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at"`
	CompletedAt *time.Time    `bson:"completed_at,omitempty"`
}

func newTodoDocument(todo *domain.Todo) (todoDocument, error) {
	if todo == nil {
		return todoDocument{}, fmt.Errorf("%w: todo is required", domain.ErrInvalidTodo)
	}

	objectID, err := objectIDFromTodoID(todo.ID)
	if err != nil {
		return todoDocument{}, err
	}

	return todoDocument{
		ID:          objectID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
		CompletedAt: copyTimePtr(todo.CompletedAt),
	}, nil
}

func (document todoDocument) toDomain() (*domain.Todo, error) {
	return domain.RehydrateTodo(
		domain.TodoID(document.ID.Hex()),
		document.Title,
		document.Description,
		document.Completed,
		document.CreatedAt,
		document.UpdatedAt,
		document.CompletedAt,
	)
}

func objectIDFromTodoID(id domain.TodoID) (bson.ObjectID, error) {
	objectID, err := bson.ObjectIDFromHex(id.String())
	if err != nil {
		return bson.NilObjectID, fmt.Errorf("%w: id must be a mongodb object id", domain.ErrInvalidTodo)
	}
	return objectID, nil
}

func copyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	copied := value.UTC()
	return &copied
}
