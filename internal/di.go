package internal

import (
	"github.com/google/wire"

	http "go-crud-db-p2/internal/adapters/primary/http/platform"
	mongoadapter "go-crud-db-p2/internal/adapters/secondary/mongo/platform"
	"go-crud-db-p2/internal/adapters/security"
	ports "go-crud-db-p2/internal/core/ports/platform"
	services "go-crud-db-p2/internal/core/services/platform"
)

var RepoSet = wire.NewSet(
	mongoadapter.NewUserRepository,
	wire.Bind(new(ports.IUserRepository), new(*mongoadapter.UserRepository)),

	mongoadapter.NewTodoRepository,
	wire.Bind(new(ports.ITodoRepository), new(*mongoadapter.TodoRepository)),

	mongoadapter.NewObjectIDGenerator,
	wire.Bind(new(ports.IUserIDGenerator), new(*mongoadapter.ObjectIDGenerator)),
	wire.Bind(new(ports.ITodoIDGenerator), new(*mongoadapter.ObjectIDGenerator)),
)

var ServiceSet = wire.NewSet(
	services.NewSystemClock,
	wire.Bind(new(ports.IClock), new(*services.SystemClock)),

	security.NewJWTManager,
	wire.Bind(new(ports.ITokenProvider), new(*security.JWTManager)),

	security.NewBcryptPasswordHasher,
	wire.Bind(new(ports.IPasswordHasher), new(*security.BcryptPasswordHasher)),

	security.NewGoogleOAuth,
	wire.Bind(new(ports.IGoogleIdentityProvider), new(*security.GoogleOAuth)),

	security.NewMemoryStateStore,
	wire.Bind(new(ports.IAuthStateStore), new(*security.MemoryStateStore)),

	services.NewAuthService,
	wire.Bind(new(ports.IAuthService), new(*services.AuthService)),

	services.NewTodoService,
	wire.Bind(new(ports.ITodoService), new(*services.TodoService)),
)

var HandlerSet = wire.NewSet(
	http.NewPlatformHandler,
)
