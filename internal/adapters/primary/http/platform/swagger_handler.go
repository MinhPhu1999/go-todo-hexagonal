package platform

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *PlatformHandler) SwaggerIndex(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(UIHTML))
}

func (h *PlatformHandler) OpenAPI(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "application/json; charset=utf-8", []byte(OpenAPIJSON))
}
