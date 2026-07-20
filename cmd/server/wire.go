//go:build wireinject
// +build wireinject

package main

import (
	"context"
	"go-crud-db-p2/config"
	"go-crud-db-p2/internal"
	platform "go-crud-db-p2/internal/adapters/primary/http/platform"

	"github.com/google/wire"
)

var CoreAppSet = wire.NewSet(
	provideGinEngine,
	provideGinRouterGroup,

	internal.ServiceSet,
	internal.HandlerSet,
)

func InitializeMongoApp(ctx context.Context, cfg config.Config) (*platform.PlatformHandler, error) {
	wire.Build(
		CoreAppSet,
		provideMongoDatabase,
		internal.MongoRepoSet,
	)
	return nil, nil
}

func InitializePostgresApp(ctx context.Context, cfg config.Config) (*platform.PlatformHandler, error) {
	wire.Build(
		CoreAppSet,
		providePostgresPool,
		internal.PostgresRepoSet,
	)
	return nil, nil
}
