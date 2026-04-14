package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger retorna um middleware Gin que registra cada requisição HTTP
// usando slog com campos estruturados.
//
// Cada entrada de log contém: método, path, status, latência, IP do cliente
// e — em caso de erro — a mensagem de erro capturada pelo Gin.
//
// Em produção os logs saem em JSON; em desenvolvimento saem legíveis no terminal.
// O formato é controlado pelo handler configurado em main.go via slog.SetDefault.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", c.ClientIP()),
		}

		if query != "" {
			attrs = append(attrs, slog.String("query", query))
		}

		if errMsg := c.Errors.ByType(gin.ErrorTypePrivate).String(); errMsg != "" {
			attrs = append(attrs, slog.String("error", errMsg))
		}

		// Nível do log baseado no status HTTP:
		// 5xx → Error, 4xx → Warn, demais → Info.
		msg := "request"
		switch {
		case status >= 500:
			slog.LogAttrs(c.Request.Context(), slog.LevelError, msg, attrs...)
		case status >= 400:
			slog.LogAttrs(c.Request.Context(), slog.LevelWarn, msg, attrs...)
		default:
			slog.LogAttrs(c.Request.Context(), slog.LevelInfo, msg, attrs...)
		}
	}
}
