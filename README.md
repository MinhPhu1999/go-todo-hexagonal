# TO-DO List API P2

MongoDB/Postgres-backed TO-DO List API with email/password auth, Google OAuth sign-in, password reset via email OTP, protected todo CRUD, Swagger docs, structured logging, and a hexagonal folder flow.

## Project Layout

```text
cmd/server                         # HTTP application entrypoint
config                             # Viper + .env/YAML config loading
db/migrations                      # MongoDB migration runner package
internal/core/domain/platform      # Auth and todo domain models
internal/core/ports/platform       # Repository/provider/service contracts
internal/core/services/platform    # Business logic
internal/adapters/primary/http/platform
                                   # Gin handler, route registration, Swagger
internal/adapters/secondary/email/platform
                                   # Console/SMTP email sender adapters
internal/adapters/secondary/mongo/platform
                                   # MongoDB repositories and client
internal/adapters/secondary/postgres/platform
                                   # PostgreSQL repositories and client
internal/adapters/security         # JWT, bcrypt, Google OAuth, OAuth state
pkg/logger                         # slog + Gin request/recovery logging
pkg/response                       # API response envelopes
```

## Configuration

`cmd/server/main.go` calls:

```go
config.LoadConfig(".", os.Getenv("ENV"))
```

Run commands from the project root so the loader can find `./.env`.

Config sources:

```text
defaults < optional ./<ENV>.yaml < .env / exported environment variables
```

Existing exported environment variables override values in `.env` because `gotenv.Load` does not replace variables that already exist.

Minimum local `.env` shape:

```bash
APP_ENV=development
GIN_MODE=debug
SERVER_ADDRESS=:8080
CONTEXT_TIMEOUT=2s

LOG_FORMAT=text
LOG_FILE_PATH=storage/logs/app.log
LOG_LEVEL=info
LOG_TO_STDOUT=true

MONGODB_URI=mongodb://localhost:27017
DB_NAME=todo_app
CONNECT_TIMEOUT=10s

JWT_SECRET=dev-secret-with-at-least-16-chars
JWT_ISSUER=go-crud-db-p2
JWT_EXPIRES_IN=24h

GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=http://localhost:5173/auth/google/callback
GOOGLE_STATE_TTL=10m

EMAIL_FROM_ADDRESS=noreply@todoapp.com
EMAIL_FROM_NAME=Todo App

OTP_LENGTH=6
OTP_EXPIRES_IN=10m
```

`JWT_SECRET` must be at least 16 characters or register/login token issuing will fail.

To enable real email delivery, also set these in `.env`:
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```
When `SMTP_HOST` is empty, OTPs are logged to the console (development mode).

## Run With Docker

This is the simplest full local run because Compose starts MongoDB and the API together.

```bash
cd /Users/crimson/Workspace/Projects/SELF/LEARN-GO/go-crud-db-p2
docker compose up --build
```

The API will be available at:

```text
http://localhost:8080
```

Stop it with:

```bash
docker compose down
```

Remove the MongoDB volume when you want a fresh database:

```bash
docker compose down -v
```

## Run Locally

Use this when MongoDB is already reachable from your machine, including a local MongoDB instance or a MongoDB Atlas URI in `.env`.

```bash
cd /Users/crimson/Workspace/Projects/SELF/LEARN-GO/go-crud-db-p2
go run ./cmd/server
```

To run the API locally against the Compose MongoDB container:

```bash
cd /Users/crimson/Workspace/Projects/SELF/LEARN-GO/go-crud-db-p2
docker compose up -d mongodb
MONGODB_URI=mongodb://localhost:27017 JWT_SECRET=dev-secret-with-at-least-16-chars go run ./cmd/server
```

## Test The API

Health check:

```bash
curl http://localhost:8080/health
```

Register a user:

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"password123","name":"Demo User"}'
```

Login:

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"password123"}'
```

Use the returned `data.access_token` for protected todo endpoints:

```bash
curl -s -X POST http://localhost:8080/api/v1/todos \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <access_token>' \
  -d '{"title":"Read the API","description":"Created from curl"}'
```

Fetch todos:

```bash
curl -s 'http://localhost:8080/api/v1/todos?pageSize=10&pageNumber=1' \
  -H 'Authorization: Bearer <access_token>'
```

Request a password reset OTP:

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com"}'
```

Reset password with OTP:

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","otp":"123456","new_password":"newSecurePass99"}'
```

## Swagger

Open:

```text
http://localhost:8080/swagger
```

The OpenAPI JSON is available at:

```text
http://localhost:8080/swagger/openapi.json
```

## Endpoints

```text
GET    /health
GET    /swagger
GET    /swagger/index.html
GET    /swagger/openapi.json

POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password
GET    /api/v1/auth/google/url
GET    /api/v1/auth/google/login
GET    /api/v1/auth/google/callback
GET    /api/v1/auth/me

POST   /api/v1/todos
GET    /api/v1/todos
GET    /api/v1/todos/:id
PATCH  /api/v1/todos/:id
DELETE /api/v1/todos/:id
```

`POST /api/v1/auth/forgot-password` and `POST /api/v1/auth/reset-password` are public endpoints. All other `/api/v1/` endpoints except auth/register, auth/login, and Google OAuth require:

```text
Authorization: Bearer <access_token>
```

## Development Checks

```bash
go test ./...
go build -o /tmp/go-crud-db-p2-server ./cmd/server
```

If the default Go build cache has local permission issues, use a writable cache:

```bash
GOCACHE=/private/tmp/go-build-cache go build -o /tmp/go-crud-db-p2-server ./cmd/server
```

## Migrations

The `db/migrations` package contains a MongoDB migration runner and migration definitions, but there is currently no `cmd/migrate` command and `cmd/server` does not run migrations automatically.
