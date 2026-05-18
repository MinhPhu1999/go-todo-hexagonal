package platform

import (
	"net/http"

	"go-crud-db-p2/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *PlatformHandler) CheckHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, response.Health())
}
