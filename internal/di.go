package internal

import (
	"github.com/google/wire"

	http "go-crud-db-p2/internal/adapters/primary/http/platform"
	emailadapter "go-crud-db-p2/internal/adapters/secondary/email/platform"
	mongoadapter "go-crud-db-p2/internal/adapters/secondary/mongo/platform"
	postgresadapter "go-crud-db-p2/internal/adapters/secondary/postgres/platform"
	security "go-crud-db-p2/internal/adapters/security"
	ports "go-crud-db-p2/internal/core/ports/platform"
	services "go-crud-db-p2/internal/core/services/platform"
)

var MongoRepoSet = wire.NewSet(
	mongoadapter.NewUserRepository,
	wire.Bind(new(ports.IUserRepository), new(*mongoadapter.UserRepository)),

	mongoadapter.NewTodoRepository,
	wire.Bind(new(ports.ITodoRepository), new(*mongoadapter.TodoRepository)),

	mongoadapter.NewObjectIDGenerator,
	wire.Bind(new(ports.IUserIDGenerator), new(*mongoadapter.ObjectIDGenerator)),
	wire.Bind(new(ports.ITodoIDGenerator), new(*mongoadapter.ObjectIDGenerator)),
)

var PostgresRepoSet = wire.NewSet(
	postgresadapter.NewUserRepository,
	wire.Bind(new(ports.IUserRepository), new(*postgresadapter.UserRepository)),

	postgresadapter.NewTodoRepository,
	wire.Bind(new(ports.ITodoRepository), new(*postgresadapter.TodoRepository)),

	postgresadapter.NewUUIDGenerator,
	wire.Bind(new(ports.IUserIDGenerator), new(*postgresadapter.UUIDGenerator)),
	wire.Bind(new(ports.ITodoIDGenerator), new(*postgresadapter.UUIDGenerator)),
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

	emailadapter.NewEmailSender,

	services.NewAuthService,
	wire.Bind(new(ports.IAuthService), new(*services.AuthService)),

	services.NewTodoService,
	wire.Bind(new(ports.ITodoService), new(*services.TodoService)),
)

var HandlerSet = wire.NewSet(
	http.NewPlatformHandler,
)
