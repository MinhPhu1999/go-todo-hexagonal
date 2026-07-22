package platform

import (
	"net/http"
	"strings"

	ports "go-crud-db-p2/internal/core/ports/platform"
	"go-crud-db-p2/pkg/response"

	"github.com/gin-gonic/gin"
)

const authUserIDKey = "auth_user_id"

type PlatformHandler struct {
	router  *gin.RouterGroup
	engine  *gin.Engine
	authSvc ports.IAuthService
	todoSvc ports.ITodoService
	tokens  ports.ITokenProvider
}

func NewPlatformHandler(
	router *gin.RouterGroup,
	engine *gin.Engine,
	authSvc ports.IAuthService,
	todoSvc ports.ITodoService,
	tokens ports.ITokenProvider,
) *PlatformHandler {
	handler := &PlatformHandler{
		router:  router,
		engine:  engine,
		authSvc: authSvc,
		todoSvc: todoSvc,
		tokens:  tokens,
	}

	handler.RegisterRoutes(handler.router)
	handler.RegisterSystemRoutes(handler.engine)
	return handler
}

func (h *PlatformHandler) Router() *gin.RouterGroup {
	return h.router
}

func (h *PlatformHandler) Engine() *gin.Engine {
	return h.engine
}

func (h *PlatformHandler) RegisterRoutes(router *gin.RouterGroup) *gin.RouterGroup {
	router.POST("/auth/register", h.Register)
	router.POST("/auth/login", h.Login)
	router.POST("/auth/forgot-password", h.ForgotPassword)
	router.POST("/auth/reset-password", h.ResetPassword)
	router.GET("/auth/google/url", h.GoogleLoginURL)
	router.GET("/auth/google/login", h.GoogleLogin)
	router.GET("/auth/google/callback", h.GoogleCallback)

	authGroup := router.Group("/")
	authGroup.Use(h.jwtMiddleware())
	authGroup.GET("/auth/me", h.Me)
	authGroup.PATCH("/auth/me", h.UpdateProfile)
	authGroup.POST("/auth/change-password", h.ChangePassword)

	authGroup.POST("/todos", h.CreateTodo)
	authGroup.GET("/todos", h.FetchTodos)
	authGroup.GET("/todos/:id", h.GetTodo)
	authGroup.PATCH("/todos/:id", h.UpdateTodo)
	authGroup.DELETE("/todos/:id", h.DeleteTodo)

	return authGroup
}

func (h *PlatformHandler) RegisterSystemRoutes(engine *gin.Engine) {
	engine.GET("/health", h.CheckHealth)
	engine.GET("/swagger", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	engine.GET("/swagger/index.html", h.SwaggerIndex)
	engine.GET("/swagger/openapi.json", h.OpenAPI)
}

func (h *PlatformHandler) jwtMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := bearerToken(ctx.GetHeader("Authorization"))
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "missing bearer token"))
			return
		}

		claims, err := h.tokens.VerifyToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "invalid bearer token"))
			return
		}

		ctx.Set(authUserIDKey, claims.UserID.String())
		ctx.Next()
	}
}

func authUserID(ctx *gin.Context) (string, bool) {
	value, ok := ctx.Get(authUserIDKey)
	if !ok {
		return "", false
	}
	userID, ok := value.(string)
	return userID, ok && userID != ""
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
