//go:build wireinject

package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"go-crud-db-p2/config"
	"go-crud-db-p2/internal"
	platform "go-crud-db-p2/internal/adapters/primary/http/platform"
	mongoadapter "go-crud-db-p2/internal/adapters/secondary/mongo/platform"
)

func InitializeApp(ctx context.Context, cfg config.Config) (*platform.PlatformHandler, error) {
	wire.Build(
		provideGinEngine,
		provideGinRouterGroup,
		provideMongoDatabase,

		provideTodoCollectionName,
		provideUserCollectionName,

		provideJWTSecret,
		provideJWTIssuer,
		provideJWTExpiresIn,

		provideGoogleConfig,
		provideGoogleStateTTL,

		provideContextTimeout,

		internal.RepoSet,
		internal.ServiceSet,
		internal.HandlerSet,
	)
	return nil, nil
}

func provideGinEngine(cfg config.Config) *gin.Engine {
	gin.SetMode(cfg.GinMode)
	engine := gin.New()
	engine.Use(CORS())
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

func provideTodoCollectionName(cfg config.Config) config.TodoCollectionName {
	return config.TodoCollectionName(cfg.Database.TodoCollection)
}

func provideUserCollectionName(cfg config.Config) config.UserCollectionName {
	return config.UserCollectionName(cfg.Database.UserCollection)
}

func provideJWTSecret(cfg config.Config) config.JWTSecret {
	return config.JWTSecret(cfg.JWT.Secret)
}

func provideJWTIssuer(cfg config.Config) config.JWTIssuer {
	return config.JWTIssuer(cfg.JWT.Issuer)
}

func provideJWTExpiresIn(cfg config.Config) config.JWTExpiresIn {
	return config.JWTExpiresIn(cfg.JWT.ExpiresIn)
}

func provideGoogleConfig(cfg config.Config) config.GoogleConfig {
	return cfg.Google
}

func provideGoogleStateTTL(cfg config.Config) config.GoogleStateTTL {
	return config.GoogleStateTTL(cfg.Google.StateTTL)
}

func provideContextTimeout(cfg config.Config) time.Duration {
	return cfg.Context.Timeout
}
