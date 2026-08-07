package response

import (
	"errors"
	"net/http"
	"path/filepath"
	"runtime"
	"stationery-management/pkg/logger"

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

func JSONError(c *gin.Context, statusCode int, message string, errorCode string, errs ...string) {
	if len(errs) == 0 {
		errs = []string{message}
	}

	// Auto-detect caller folder, file, and handler function name for unique error code (e.g. UKJ001)
	folder := "handler"
	file := "handler"
	fnName := "Handler"
	pc, fileP, _, ok := runtime.Caller(2)
	if ok {
		file = filepath.Base(fileP)
		dir := filepath.Dir(fileP)
		folder = filepath.Base(dir)
		if fn := runtime.FuncForPC(pc); fn != nil {
			fnName = fn.Name()
		}
	}

	uniqueCode := logger.LogWithCode(folder, file, fnName, errors.New(message))
	if errorCode == "" || errorCode == "BAD_REQUEST" || errorCode == "INTERNAL_SERVER_ERROR" {
		errorCode = uniqueCode
	}

	c.JSON(statusCode, APIResponse{
		Success:   false,
		Message:   message,
		ErrorCode: errorCode,
		Errors:    errs,
	})
}

func BadRequest(c *gin.Context, message string, errs ...string) {
	JSONError(c, http.StatusBadRequest, message, "", errs...)
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
	JSONError(c, http.StatusInternalServerError, message, "")
}
