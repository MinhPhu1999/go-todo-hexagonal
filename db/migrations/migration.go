package migrations

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const CollectionName = "schema_migrations"

type Config struct {
	TodoCollection string
	UserCollection string
}

type Migration struct {
	Version string
	Name    string
	Up      func(context.Context, *mongo.Database, Config) error
	Down    func(context.Context, *mongo.Database, Config) error
}

type Status struct {
	Version   string
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

type Runner struct {
	database   *mongo.Database
	config     Config
	collection *mongo.Collection
	migrations []Migration
	logger     *slog.Logger
}

func NewRunner(database *mongo.Database, config Config, migrations []Migration, logger *slog.Logger) *Runner {
	ordered := append([]Migration(nil), migrations...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Version < ordered[j].Version
	})

	if logger == nil {
		logger = slog.Default()
	}

	return &Runner{
		database:   database,
		config:     config,
		collection: database.Collection(CollectionName),
		migrations: ordered,
		logger:     logger,
	}
}

func (runner *Runner) Up(ctx context.Context) error {
	if err := runner.validate(); err != nil {
		return err
	}
	if err := runner.ensureCollection(ctx); err != nil {
		return err
	}

	applied, err := runner.appliedRecords(ctx)
	if err != nil {
		return err
	}

	for _, item := range runner.migrations {
		if _, ok := applied[item.Version]; ok {
			runner.logger.Info("migration skipped", "version", item.Version, "name", item.Name)
			continue
		}

		runner.logger.Info("migration started", "direction", "up", "version", item.Version, "name", item.Name)
		if err := item.Up(ctx, runner.database, runner.config); err != nil {
			return fmt.Errorf("run migration %s up: %w", item.Version, err)
		}

		now := time.Now().UTC()
		record := migrationRecord{
			ID:        item.Version,
			Version:   item.Version,
			Name:      item.Name,
			AppliedAt: now,
		}

		if _, err := runner.collection.InsertOne(ctx, record); err != nil {
			return fmt.Errorf("record migration %s: %w", item.Version, err)
		}
		runner.logger.Info("migration completed", "direction", "up", "version", item.Version, "name", item.Name)
	}

	return nil
}

func (runner *Runner) Down(ctx context.Context, steps int) error {
	if steps <= 0 {
		steps = 1
	}
	if err := runner.validate(); err != nil {
		return err
	}
	if err := runner.ensureCollection(ctx); err != nil {
		return err
	}

	applied, err := runner.appliedRecordsByAppliedAtDesc(ctx)
	if err != nil {
		return err
	}
	if len(applied) == 0 {
		runner.logger.Info("no applied migrations to roll back")
		return nil
	}
	if steps > len(applied) {
		steps = len(applied)
	}

	byVersion := runner.migrationsByVersion()
	for _, record := range applied[:steps] {
		item, ok := byVersion[record.Version]
		if !ok {
			return fmt.Errorf("applied migration %s is missing from code", record.Version)
		}

		runner.logger.Info("migration started", "direction", "down", "version", item.Version, "name", item.Name)
		if err := item.Down(ctx, runner.database, runner.config); err != nil {
			return fmt.Errorf("run migration %s down: %w", item.Version, err)
		}

		if _, err := runner.collection.DeleteOne(ctx, bson.M{"_id": record.Version}); err != nil {
			return fmt.Errorf("delete migration record %s: %w", record.Version, err)
		}
		runner.logger.Info("migration completed", "direction", "down", "version", item.Version, "name", item.Name)
	}

	return nil
}

func (runner *Runner) Status(ctx context.Context) ([]Status, error) {
	if err := runner.validate(); err != nil {
		return nil, err
	}
	if err := runner.ensureCollection(ctx); err != nil {
		return nil, err
	}

	applied, err := runner.appliedRecords(ctx)
	if err != nil {
		return nil, err
	}

	statuses := make([]Status, 0, len(runner.migrations))
	for _, item := range runner.migrations {
		status := Status{
			Version: item.Version,
			Name:    item.Name,
		}

		if record, ok := applied[item.Version]; ok {
			appliedAt := record.AppliedAt
			status.Applied = true
			status.AppliedAt = &appliedAt
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (runner *Runner) ensureCollection(ctx context.Context) error {
	indexes := runner.collection.Indexes()
	_, err := indexes.CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "applied_at", Value: 1}},
		Options: options.Index().SetName("idx_schema_migrations_applied_at"),
	})
	if err != nil {
		return fmt.Errorf("ensure migration indexes: %w", err)
	}

	return nil
}

func (runner *Runner) appliedRecords(ctx context.Context) (map[string]migrationRecord, error) {
	cursor, err := runner.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("find applied migrations: %w", err)
	}
	defer cursor.Close(ctx)

	records := make(map[string]migrationRecord)
	for cursor.Next(ctx) {
		var record migrationRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, fmt.Errorf("decode migration record: %w", err)
		}
		records[record.Version] = record
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}

	return records, nil
}

func (runner *Runner) appliedRecordsByAppliedAtDesc(ctx context.Context) ([]migrationRecord, error) {
	cursor, err := runner.collection.Find(
		ctx,
		bson.D{},
		options.Find().SetSort(bson.D{
			{Key: "applied_at", Value: -1},
			{Key: "version", Value: -1},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("find applied migrations: %w", err)
	}
	defer cursor.Close(ctx)

	records := make([]migrationRecord, 0)
	for cursor.Next(ctx) {
		var record migrationRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, fmt.Errorf("decode migration record: %w", err)
		}
		records = append(records, record)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}

	return records, nil
}

func (runner *Runner) migrationsByVersion() map[string]Migration {
	byVersion := make(map[string]Migration, len(runner.migrations))
	for _, item := range runner.migrations {
		byVersion[item.Version] = item
	}
	return byVersion
}

func (runner *Runner) validate() error {
	seen := make(map[string]struct{}, len(runner.migrations))
	for _, item := range runner.migrations {
		if item.Version == "" {
			return fmt.Errorf("migration version is required")
		}
		if item.Name == "" {
			return fmt.Errorf("migration %s name is required", item.Version)
		}
		if item.Up == nil {
			return fmt.Errorf("migration %s up function is required", item.Version)
		}
		if item.Down == nil {
			return fmt.Errorf("migration %s down function is required", item.Version)
		}
		if _, ok := seen[item.Version]; ok {
			return fmt.Errorf("duplicate migration version %s", item.Version)
		}
		seen[item.Version] = struct{}{}
	}

	return nil
}

type migrationRecord struct {
	ID        string    `bson:"_id"`
	Version   string    `bson:"version"`
	Name      string    `bson:"name"`
	AppliedAt time.Time `bson:"applied_at"`
}
