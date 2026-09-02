package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {

		log.Error().
			Interface("error", recovered).
			Str("service", "invoicing").
			Str("operation", "http_request").
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Msg("panic recuperado")

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Erro interno do servidor",
		})
	})
}
