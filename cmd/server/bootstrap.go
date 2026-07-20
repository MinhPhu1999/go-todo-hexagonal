package main

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"go-crud-db-p2/config"
	middleware "go-crud-db-p2/internal/adapters/primary/http/middleware"
	platform "go-crud-db-p2/internal/adapters/primary/http/platform"
	mongoadapter "go-crud-db-p2/internal/adapters/secondary/mongo/platform"
	postgresadapter "go-crud-db-p2/internal/adapters/secondary/postgres/platform"
	"go-crud-db-p2/pkg/logger"
)

func BootstrapApp(ctx context.Context, cfg config.Config) (*platform.PlatformHandler, error) {
	switch cfg.Database.Driver {
	case "postgres":
		return InitializePostgresApp(ctx, cfg)
	default:
		return InitializeMongoApp(ctx, cfg)
	}
}

func provideGinEngine(cfg config.Config) *gin.Engine {
	gin.SetMode(cfg.GinMode)
	engine := gin.New()
	engine.Use(logger.GinRecovery(slog.Default()))
	engine.Use(logger.GinRequestLogger(slog.Default()))
	engine.Use(middleware.CORS())
	return engine
}

func provideGinRouterGroup(engine *gin.Engine) *gin.RouterGroup {
	return engine.Group("/api/v1")
}

func provideMongoDatabase(ctx context.Context, cfg config.Config) (*mongo.Database, error) {
	client, err := mongoadapter.Connect(ctx, cfg.Database.URI)
	if err != nil {
		return nil, err
	}
	return client.Database(cfg.Database.Name), nil
}

func providePostgresPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	return postgresadapter.Connect(ctx, cfg.Database.PostgresDSN)
}
