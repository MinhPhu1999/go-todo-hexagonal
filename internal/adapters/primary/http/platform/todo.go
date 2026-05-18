package platform

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	domain "go-crud-db-p2/internal/core/domain/platform"
	"go-crud-db-p2/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *PlatformHandler) CreateTodo(ctx *gin.Context) {
	var request domain.CreateTodoRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error("BAD_REQUEST", "invalid json request body"))
		return
	}

	todo, err := h.todoSvc.Create(ctx.Request.Context(), request)
	if err != nil {
		handleTodoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, response.Created(todo))
}

func (h *PlatformHandler) FetchTodos(ctx *gin.Context) {
	request, err := parseFetchTodosRequest(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error("BAD_REQUEST", err.Error()))
		return
	}

	todoList, err := h.todoSvc.Fetch(ctx.Request.Context(), request)
	if err != nil {
		handleTodoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.OKWithMeta(todoList.Items, response.Meta{
		Page:       todoList.PageNumber,
		Limit:      todoList.PageSize,
		Total:      int(todoList.Total),
		TotalPages: todoList.TotalPages(),
	}))
}

func (h *PlatformHandler) GetTodo(ctx *gin.Context) {
	todo, err := h.todoSvc.GetByID(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		handleTodoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.OK(todo))
}

func (h *PlatformHandler) UpdateTodo(ctx *gin.Context) {
	var request domain.UpdateTodoRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error("BAD_REQUEST", "invalid json request body"))
		return
	}

	todo, err := h.todoSvc.Update(ctx.Request.Context(), ctx.Param("id"), request)
	if err != nil {
		handleTodoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.OKWithMessage(todo, "updated"))
}

func (h *PlatformHandler) DeleteTodo(ctx *gin.Context) {
	if err := h.todoSvc.Delete(ctx.Request.Context(), ctx.Param("id")); err != nil {
		handleTodoError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.Deleted())
}

func handleTodoError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidTodo):
		ctx.JSON(http.StatusBadRequest, response.Error("INVALID_TODO", err.Error()))
	case errors.Is(err, domain.ErrTodoNotFound):
		ctx.JSON(http.StatusNotFound, response.Error("TODO_NOT_FOUND", "todo not found"))
	default:
		ctx.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", "internal server error"))
	}
}

func parseFetchTodosRequest(ctx *gin.Context) (domain.FetchTodosRequest, error) {
	var request domain.FetchTodosRequest

	pageSize, err := positiveIntQuery(ctx, "pageSize")
	if err != nil {
		return domain.FetchTodosRequest{}, err
	}
	request.PageSize = pageSize

	pageNumber, err := positiveIntQuery(ctx, "pageNumber")
	if err != nil {
		return domain.FetchTodosRequest{}, err
	}
	request.PageNumber = pageNumber

	if value, ok := ctx.GetQuery("completed"); ok {
		completed, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return domain.FetchTodosRequest{}, fmt.Errorf("completed must be a boolean")
		}
		request.Completed = &completed
	}

	if value, ok := ctx.GetQuery("search"); ok {
		request.Search = value
	}

	return request, nil
}

func positiveIntQuery(ctx *gin.Context, name string) (int, error) {
	value, ok := ctx.GetQuery(name)
	if !ok {
		return 0, nil
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}

	return parsed, nil
}
