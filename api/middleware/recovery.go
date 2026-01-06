package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery returns a recovery middleware
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("Panic recovered",
			zap.Any("error", recovered),
			zap.String("path", c.Request.URL.Path),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_server_error",
			"message": "An internal error occurred",
		})
		c.Abort()
	})
}

