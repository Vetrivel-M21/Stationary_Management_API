package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	ErrorCode string      `json:"errorCode,omitempty"`
	Errors    []string    `json:"errors,omitempty"`
}

func JSONSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func JSONError(c *gin.Context, statusCode int, message string, errorCode string, errors ...string) {
	if len(errors) == 0 {
		errors = []string{message}
	}
	c.JSON(statusCode, APIResponse{
		Success:   false,
		Message:   message,
		ErrorCode: errorCode,
		Errors:    errors,
	})
}

func BadRequest(c *gin.Context, message string, errs ...string) {
	JSONError(c, http.StatusBadRequest, message, "BAD_REQUEST", errs...)
}

func Unauthorized(c *gin.Context, message string) {
	JSONError(c, http.StatusUnauthorized, message, "UNAUTHORIZED")
}

func Forbidden(c *gin.Context, message string) {
	JSONError(c, http.StatusForbidden, message, "FORBIDDEN")
}

func NotFound(c *gin.Context, message string) {
	JSONError(c, http.StatusNotFound, message, "NOT_FOUND")
}

func InternalError(c *gin.Context, message string) {
	JSONError(c, http.StatusInternalServerError, message, "INTERNAL_SERVER_ERROR")
}
