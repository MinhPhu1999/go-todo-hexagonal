package platform

const OpenAPIJSON = `{
  "openapi": "3.0.3",
  "info": {
    "title": "TO-DO List API",
    "description": "MongoDB-backed TO-DO List API with email/password and Google sign-in.",
    "version": "1.1.0"
  },
  "servers": [
    {
      "url": "/",
      "description": "Current server"
    }
  ],
  "tags": [
    {
      "name": "Health",
      "description": "Application health"
    },
    {
      "name": "Auth",
      "description": "Registration, login, Google sign-in, and current user"
    },
    {
      "name": "Todos",
      "description": "TO-DO List operations"
    }
  ],
  "paths": {
    "/health": {
      "get": {
        "tags": ["Health"],
        "summary": "Check API health",
        "responses": {
          "200": {
            "description": "API is healthy",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/HealthResponse" },
                "example": { "success": true, "message": "ok" }
              }
            }
          }
        }
      }
    },
    "/api/v1/auth/register": {
      "post": {
        "tags": ["Auth"],
        "summary": "Register with email and password",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/RegisterRequest" },
              "example": {
                "email": "demo@example.com",
                "password": "password123",
                "name": "Demo User"
              }
            }
          }
        },
        "responses": {
          "201": { "$ref": "#/components/responses/AuthSuccess" },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "409": { "$ref": "#/components/responses/Conflict" },
          "500": { "$ref": "#/components/responses/InternalServerError" }
        }
      }
    },
    "/api/v1/auth/login": {
      "post": {
        "tags": ["Auth"],
        "summary": "Sign in with email and password",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/LoginRequest" },
              "example": {
                "email": "demo@example.com",
                "password": "password123"
              }
            }
          }
        },
        "responses": {
          "200": { "$ref": "#/components/responses/AuthSuccess" },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "500": { "$ref": "#/components/responses/InternalServerError" }
        }
      }
    },
    "/api/v1/auth/google/url": {
      "get": {
        "tags": ["Auth"],
        "summary": "Get Google sign-in URL",
        "description": "Use this endpoint from Swagger. Copy the returned data.url and open it in the browser. The redirect endpoint /api/v1/auth/google/login is for direct browser navigation and is not used by Swagger because browser fetch blocks cross-origin OAuth redirects.",
        "responses": {
          "200": {
            "description": "Google OAuth URL",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/GoogleLoginURLResponse" }
              }
            }
          },
          "503": {
            "$ref": "#/components/responses/GoogleUnavailable"
          }
        }
      }
    },
    "/api/v1/auth/google/callback": {
      "get": {
        "tags": ["Auth"],
        "summary": "Google OAuth callback",
        "parameters": [
          {
            "name": "state",
            "in": "query",
            "required": true,
            "schema": { "type": "string" }
          },
          {
            "name": "code",
            "in": "query",
            "required": true,
            "schema": { "type": "string" }
          }
        ],
        "responses": {
          "200": { "$ref": "#/components/responses/AuthSuccess" },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "502": { "$ref": "#/components/responses/GoogleProfileUnavailable" },
          "503": { "$ref": "#/components/responses/GoogleUnavailable" }
        }
      }
    },
    "/api/v1/auth/me": {
      "get": {
        "tags": ["Auth"],
        "summary": "Get current signed-in user",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "Current user",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/UserResponse" }
              }
            }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" }
        }
      }
    },
    "/api/v1/todos": {
      "post": {
        "tags": ["Todos"],
        "summary": "Create a todo",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/CreateTodoRequest" },
              "example": {
                "title": "Learn Swagger",
                "description": "Call the API from the browser."
              }
            }
          }
        },
        "responses": {
          "201": { "$ref": "#/components/responses/TodoSuccess" },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "500": { "$ref": "#/components/responses/InternalServerError" }
        }
      },
      "get": {
        "tags": ["Todos"],
        "summary": "List todos",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          {
            "name": "pageSize",
            "in": "query",
            "required": false,
            "description": "Number of todos per page",
            "schema": { "type": "integer", "minimum": 1, "maximum": 100, "default": 10, "example": 10 }
          },
          {
            "name": "pageNumber",
            "in": "query",
            "required": false,
            "description": "Page number to return",
            "schema": { "type": "integer", "minimum": 1, "default": 1, "example": 1 }
          },
          {
            "name": "completed",
            "in": "query",
            "required": false,
            "description": "Filter todos by completion status",
            "schema": { "type": "boolean", "example": true }
          },
          {
            "name": "search",
            "in": "query",
            "required": false,
            "description": "Case-insensitive search text matched against title and description",
            "schema": { "type": "string", "example": "api" }
          }
        ],
        "responses": {
          "200": {
            "description": "Todo list",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/TodoListResponse" }
              }
            }
          },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "500": { "$ref": "#/components/responses/InternalServerError" }
        }
      }
    },
    "/api/v1/todos/{id}": {
      "get": {
        "tags": ["Todos"],
        "summary": "Get a todo by id",
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "$ref": "#/components/parameters/TodoID" }],
        "responses": {
          "200": { "$ref": "#/components/responses/TodoSuccess" },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "404": { "$ref": "#/components/responses/NotFound" },
          "500": { "$ref": "#/components/responses/InternalServerError" }
        }
      },
      "patch": {
        "tags": ["Todos"],
        "summary": "Update a todo",
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "$ref": "#/components/parameters/TodoID" }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/UpdateTodoRequest" },
              "example": { "completed": true }
            }
          }
        },
        "responses": {
          "200": { "$ref": "#/components/responses/TodoSuccess" },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "404": { "$ref": "#/components/responses/NotFound" },
          "500": { "$ref": "#/components/responses/InternalServerError" }
        }
      },
      "delete": {
        "tags": ["Todos"],
        "summary": "Delete a todo",
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "$ref": "#/components/parameters/TodoID" }],
        "responses": {
          "200": { "$ref": "#/components/responses/Deleted" },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "404": { "$ref": "#/components/responses/NotFound" },
          "500": { "$ref": "#/components/responses/InternalServerError" }
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    },
    "parameters": {
      "TodoID": {
        "name": "id",
        "in": "path",
        "required": true,
        "description": "MongoDB ObjectID hex string",
        "schema": {
          "type": "string",
          "example": "665000000000000000000101"
        }
      }
    },
    "responses": {
      "AuthSuccess": {
        "description": "Authentication succeeded",
        "content": {
          "application/json": {
            "schema": { "$ref": "#/components/schemas/AuthResponseEnvelope" }
          }
        }
      },
      "TodoSuccess": {
        "description": "Todo response",
        "content": {
          "application/json": {
            "schema": { "$ref": "#/components/schemas/TodoResponse" }
          }
        }
      },
      "Deleted": {
        "description": "Deleted",
        "content": {
          "application/json": {
            "schema": { "$ref": "#/components/schemas/DeletedResponse" },
            "example": { "success": true, "message": "deleted" }
          }
        }
      },
      "BadRequest": {
        "description": "Invalid request",
        "content": {
          "application/json": {
            "schema": { "$ref": "#/components/schemas/ErrorResponse" },
            "example": {
              "success": false,
              "error": { "code": "BAD_REQUEST", "message": "invalid json request body" }
            }
          }
        }
      },
      "Unauthorized": {
        "description": "Unauthorized",
        "content": {
          "application/json": {
            "schema": { "$ref": "#/components/schemas/ErrorResponse" },
            "example": {
              "success": false,
              "error": { "code": "UNAUTHORIZED", "message": "invalid bearer token" }
            }
          }
        }
      },
      "Conflict": {
        "description": "Conflict",
        "content": {
          "application/json": {
            "schema": { "$ref": "#/components/schemas/ErrorResponse" },
            "example": {
              "success": false,
              "error": { "code": "EMAIL_ALREADY_EXISTS", "message": "email already exists" }
            }
          }
        }
      },
      "NotFound": {
        "description": "Todo not found",
        "content": {
          "application/json": {
            "schema": { "$ref": "#/components/schemas/ErrorResponse" },
            "example": {
              "success": false,
              "error": { "code": "TODO_NOT_FOUND", "message": "todo not found" }
            }
          }
        }
      },
      "GoogleUnavailable": {
        "description": "Google sign-in not configured",
        "content": {
          "application/json": {
            "schema": { "$ref": "#/components/schemas/ErrorResponse" },
            "example": {
              "success": false,
              "error": { "code": "GOOGLE_AUTH_UNAVAILABLE", "message": "google sign in is not configured" }
            }
          }
        }
      },
      "GoogleProfileUnavailable": {
        "description": "Could not read Google profile",
        "content": {
          "application/json": {
            "schema": { "$ref": "#/components/schemas/ErrorResponse" },
            "example": {
              "success": false,
              "error": { "code": "GOOGLE_PROFILE_UNAVAILABLE", "message": "could not read google profile" }
            }
          }
        }
      },
      "InternalServerError": {
        "description": "Unexpected server error",
        "content": {
          "application/json": {
            "schema": { "$ref": "#/components/schemas/ErrorResponse" },
            "example": {
              "success": false,
              "error": { "code": "INTERNAL_SERVER_ERROR", "message": "internal server error" }
            }
          }
        }
      }
    },
    "schemas": {
      "RegisterRequest": {
        "type": "object",
        "required": ["email", "password"],
        "properties": {
          "email": { "type": "string", "format": "email", "example": "demo@example.com" },
          "password": { "type": "string", "minLength": 8, "maxLength": 72, "example": "password123" },
          "name": { "type": "string", "example": "Demo User" }
        }
      },
      "LoginRequest": {
        "type": "object",
        "required": ["email", "password"],
        "properties": {
          "email": { "type": "string", "format": "email", "example": "demo@example.com" },
          "password": { "type": "string", "example": "password123" }
        }
      },
      "CreateTodoRequest": {
        "type": "object",
        "required": ["title"],
        "properties": {
          "title": { "type": "string", "maxLength": 160, "example": "Learn Swagger" },
          "description": { "type": "string", "maxLength": 2000, "example": "Call the API from the browser." }
        }
      },
      "UpdateTodoRequest": {
        "type": "object",
        "minProperties": 1,
        "properties": {
          "title": { "type": "string", "maxLength": 160, "example": "Updated todo title" },
          "description": { "type": "string", "maxLength": 2000, "example": "Updated description" },
          "completed": { "type": "boolean", "example": true }
        }
      },
      "User": {
        "type": "object",
        "properties": {
          "id": { "type": "string", "example": "665000000000000000000201" },
          "email": { "type": "string", "format": "email", "example": "demo@example.com" },
          "name": { "type": "string", "example": "Demo User" },
          "picture": { "type": "string", "example": "https://example.com/avatar.png" },
          "providers": { "type": "array", "items": { "type": "string" }, "example": ["email"] },
          "created_at": { "type": "string", "format": "date-time" },
          "updated_at": { "type": "string", "format": "date-time" },
          "last_login_at": { "type": "string", "format": "date-time", "nullable": true }
        }
      },
      "AuthPayload": {
        "type": "object",
        "properties": {
          "token_type": { "type": "string", "example": "Bearer" },
          "access_token": { "type": "string", "example": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." },
          "expires_at": { "type": "string", "format": "date-time" },
          "user": { "$ref": "#/components/schemas/User" }
        }
      },
      "Todo": {
        "type": "object",
        "properties": {
          "id": { "type": "string", "example": "665000000000000000000101" },
          "title": { "type": "string", "example": "Review clean architecture layers" },
          "description": { "type": "string", "example": "Walk through domain, usecase, repository, controller, and route packages." },
          "completed": { "type": "boolean", "example": false },
          "created_at": { "type": "string", "format": "date-time", "example": "2026-05-12T02:00:00Z" },
          "updated_at": { "type": "string", "format": "date-time", "example": "2026-05-12T02:00:00Z" },
          "completed_at": { "type": "string", "format": "date-time", "nullable": true, "example": "2026-05-12T03:00:00Z" }
        }
      },
      "AuthResponseEnvelope": {
        "type": "object",
        "properties": {
          "success": { "type": "boolean", "example": true },
          "message": { "type": "string", "example": "signed in" },
          "data": { "$ref": "#/components/schemas/AuthPayload" }
        }
      },
      "UserResponse": {
        "type": "object",
        "properties": {
          "success": { "type": "boolean", "example": true },
          "data": { "$ref": "#/components/schemas/User" }
        }
      },
      "GoogleLoginURLResponse": {
        "type": "object",
        "properties": {
          "success": { "type": "boolean", "example": true },
          "message": { "type": "string", "example": "google login url" },
          "data": {
            "type": "object",
            "properties": {
              "url": {
                "type": "string",
                "example": "https://accounts.google.com/o/oauth2/auth?client_id=..."
              }
            }
          }
        }
      },
      "HealthResponse": {
        "type": "object",
        "properties": {
          "success": { "type": "boolean", "example": true },
          "message": { "type": "string", "example": "ok" }
        }
      },
      "TodoResponse": {
        "type": "object",
        "properties": {
          "success": { "type": "boolean", "example": true },
          "message": { "type": "string", "example": "created" },
          "data": { "$ref": "#/components/schemas/Todo" }
        }
      },
      "TodoListResponse": {
        "type": "object",
        "properties": {
          "success": { "type": "boolean", "example": true },
          "data": {
            "type": "array",
            "items": { "$ref": "#/components/schemas/Todo" }
          },
          "meta": { "$ref": "#/components/schemas/PaginationMeta" }
        }
      },
      "PaginationMeta": {
        "type": "object",
        "properties": {
          "page": { "type": "integer", "example": 1 },
          "limit": { "type": "integer", "example": 10 },
          "total": { "type": "integer", "example": 5 },
          "total_pages": { "type": "integer", "example": 1 }
        }
      },
      "DeletedResponse": {
        "type": "object",
        "properties": {
          "success": { "type": "boolean", "example": true },
          "message": { "type": "string", "example": "deleted" }
        }
      },
      "ErrorDetail": {
        "type": "object",
        "properties": {
          "code": { "type": "string", "example": "UNAUTHORIZED" },
          "message": { "type": "string", "example": "invalid bearer token" },
          "details": { "type": "string" }
        }
      },
      "ErrorResponse": {
        "type": "object",
        "properties": {
          "success": { "type": "boolean", "example": false },
          "error": { "$ref": "#/components/schemas/ErrorDetail" }
        }
      }
    }
  }
}`

const UIHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>TO-DO List API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
    <style>
      body {
        margin: 0;
        background: #f7f8fa;
      }
      .topbar {
        display: none;
      }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = function () {
        window.ui = SwaggerUIBundle({
          url: "/swagger/openapi.json",
          dom_id: "#swagger-ui",
          deepLinking: true,
          displayRequestDuration: true,
          persistAuthorization: true
        });
      };
    </script>
  </body>
</html>`
