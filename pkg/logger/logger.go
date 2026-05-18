package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"go-crud-db-p2/pkg/response"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Format    string
	FilePath  string
	Level     string
	ToStdout  bool
	AddSource bool
}

type Logger struct {
	*slog.Logger
	file *os.File
}

func New(config Config) (*Logger, error) {
	if config.FilePath == "" {
		config.FilePath = "storage/logs/app.log"
	}

	if err := os.MkdirAll(filepath.Dir(config.FilePath), 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	file, err := os.OpenFile(config.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	writer := io.Writer(file)
	if config.ToStdout {
		writer = io.MultiWriter(os.Stdout, file)
	}

	options := &slog.HandlerOptions{
		AddSource: config.AddSource,
		Level:     parseLevel(config.Level),
	}

	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(config.Format)) {
	case "json":
		handler = slog.NewJSONHandler(writer, options)
	default:
		handler = slog.NewTextHandler(writer, options)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	log.SetOutput(writer)
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)

	gin.DefaultWriter = writer
	gin.DefaultErrorWriter = writer

	return &Logger{
		Logger: logger,
		file:   file,
	}, nil
}

func (logger *Logger) Close() error {
	if logger == nil || logger.file == nil {
		return nil
	}
	return logger.file.Close()
}

func GinRequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startedAt := time.Now()

		ctx.Next()

		latency := time.Since(startedAt)
		status := ctx.Writer.Status()
		path := ctx.Request.URL.Path
		if ctx.Request.URL.RawQuery != "" {
			path += "?" + ctx.Request.URL.RawQuery
		}

		attrs := []any{
			"status", status,
			"method", ctx.Request.Method,
			"path", path,
			"client_ip", ctx.ClientIP(),
			"latency", latency.String(),
			"user_agent", ctx.Request.UserAgent(),
			"request_id", ctx.Writer.Header().Get("X-Request-ID"),
		}

		if len(ctx.Errors) > 0 {
			attrs = append(attrs, "errors", ctx.Errors.String())
		}

		switch {
		case status >= http.StatusInternalServerError:
			logger.ErrorContext(ctx.Request.Context(), "http request completed", attrs...)
		case status >= http.StatusBadRequest:
			logger.WarnContext(ctx.Request.Context(), "http request completed", attrs...)
		default:
			logger.InfoContext(ctx.Request.Context(), "http request completed", attrs...)
		}
	}
}

func GinRecovery(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(ctx *gin.Context, recovered any) {
		logger.ErrorContext(
			context.Background(),
			"panic recovered",
			"method", ctx.Request.Method,
			"path", ctx.Request.URL.Path,
			"request_id", ctx.Writer.Header().Get("X-Request-ID"),
			"error", fmt.Sprint(recovered),
			"stack", string(debug.Stack()),
		)
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", "internal server error"),
		)
	})
}

func parseLevel(value string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
