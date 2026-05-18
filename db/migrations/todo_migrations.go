package migrations

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const seedTodoID = "665000000000000000000001"

var All = []Migration{
	{
		Version: "202605120001",
		Name:    "create_todos_created_at_index",
		Up:      createTodosCreatedAtIndex,
		Down:    dropTodosCreatedAtIndex,
	},
	{
		Version: "202605120002",
		Name:    "insert_seed_todo",
		Up:      insertSeedTodo,
		Down:    deleteSeedTodo,
	},
	{
		Version: "202605120003",
		Name:    "edit_seed_todo_description",
		Up:      editSeedTodoDescription,
		Down:    restoreSeedTodoDescription,
	},
	{
		Version: "202605120004",
		Name:    "insert_mock_todos",
		Up:      insertMockTodos,
		Down:    deleteMockTodos,
	},
	{
		Version: "202605120005",
		Name:    "create_users_auth_indexes",
		Up:      createUsersAuthIndexes,
		Down:    dropUsersAuthIndexes,
	},
}

func createTodosCreatedAtIndex(ctx context.Context, database *mongo.Database, config Config) error {
	_, err := database.Collection(config.TodoCollection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "created_at", Value: -1}},
		Options: options.Index().SetName("idx_todos_created_at_desc"),
	})
	return err
}

func dropTodosCreatedAtIndex(ctx context.Context, database *mongo.Database, config Config) error {
	return database.Collection(config.TodoCollection).Indexes().DropOne(ctx, "idx_todos_created_at_desc")
}

func insertSeedTodo(ctx context.Context, database *mongo.Database, config Config) error {
	objectID, err := bson.ObjectIDFromHex(seedTodoID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err = database.Collection(config.TodoCollection).UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$setOnInsert": bson.M{
				"_id":          objectID,
				"title":        "Try the migration command",
				"description":  "This todo was inserted by a MongoDB migration.",
				"completed":    false,
				"created_at":   now,
				"updated_at":   now,
				"completed_at": nil,
			},
		},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func deleteSeedTodo(ctx context.Context, database *mongo.Database, config Config) error {
	objectID, err := bson.ObjectIDFromHex(seedTodoID)
	if err != nil {
		return err
	}

	_, err = database.Collection(config.TodoCollection).DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func editSeedTodoDescription(ctx context.Context, database *mongo.Database, config Config) error {
	objectID, err := bson.ObjectIDFromHex(seedTodoID)
	if err != nil {
		return err
	}

	_, err = database.Collection(config.TodoCollection).UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"description": "This todo was inserted and edited by MongoDB migrations.",
				"updated_at":  time.Now().UTC(),
			},
		},
	)
	return err
}

func restoreSeedTodoDescription(ctx context.Context, database *mongo.Database, config Config) error {
	objectID, err := bson.ObjectIDFromHex(seedTodoID)
	if err != nil {
		return err
	}

	_, err = database.Collection(config.TodoCollection).UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"description": "This todo was inserted by a MongoDB migration.",
				"updated_at":  time.Now().UTC(),
			},
		},
	)
	return err
}

func createUsersAuthIndexes(ctx context.Context, database *mongo.Database, config Config) error {
	indexes := database.Collection(config.UserCollection).Indexes()
	_, err := indexes.CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "email", Value: 1}},
			Options: options.Index().
				SetName("idx_users_email_unique").
				SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "google_id", Value: 1}},
			Options: options.Index().
				SetName("idx_users_google_id_unique").
				SetUnique(true).
				SetSparse(true),
		},
	})
	return err
}

func dropUsersAuthIndexes(ctx context.Context, database *mongo.Database, config Config) error {
	indexes := database.Collection(config.UserCollection).Indexes()
	if err := indexes.DropOne(ctx, "idx_users_google_id_unique"); err != nil {
		return err
	}
	return indexes.DropOne(ctx, "idx_users_email_unique")
}
