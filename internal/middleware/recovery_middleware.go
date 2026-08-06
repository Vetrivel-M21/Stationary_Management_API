package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"stationery-management/pkg/response"

	"github.com/gin-gonic/gin"
)

func PanicRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERED] %v\nStack Trace:\n%s\n", err, string(debug.Stack()))
				response.JSONError(c, http.StatusInternalServerError, "An unexpected internal server error occurred", "INTERNAL_SERVER_ERROR")
				c.Abort()
			}
		}()
		c.Next()
	}
}
