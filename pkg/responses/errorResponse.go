package responses

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// NewError создает стандартный error response
func NewErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, ErrorResponse{Error: message})
}

// NotFoundWithMessage создает кастомный 404 response
func NotFound(c *gin.Context, message string) {
	NewErrorResponse(c, http.StatusNotFound, message)
}

// Forbidden создает response с статусом 403 Forbidden
func Forbidden(c *gin.Context, message string) {
	NewErrorResponse(c, http.StatusForbidden, message)
}

// Conflict создает response с статусом 409 Conflict
func Conflict(c *gin.Context, message string) {
	NewErrorResponse(c, http.StatusConflict, message)
}

// AlreadyExists создает response для конфликта существования ресурса
func AlreadyExists(c *gin.Context, resource string) {
	Conflict(c, fmt.Sprintf("%s already exists", resource))
}

// BadRequest создает response с статусом 400 Bad Request
func BadRequest(c *gin.Context, message string) {
	NewErrorResponse(c, http.StatusBadRequest, message)
}

// Unauthorized создает response с статусом 401 Unauthorized
func Unauthorized(c *gin.Context) {
	NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
}

// UnauthorizedWithMessage создает кастомный 401 response
func UnauthorizedWithMessage(c *gin.Context, message string) {
	NewErrorResponse(c, http.StatusUnauthorized, message)
}

// InternalServerError создает response с статусом 500 Internal Server Error
func InternalServerErrorWithDetails(c *gin.Context, message string) {
	NewErrorResponse(c, http.StatusInternalServerError, message)
}

// TooManyRequests создает response с статусом 429 Too Many Requests
func TooManyRequests(c *gin.Context) {
	NewErrorResponse(c, http.StatusTooManyRequests, "Too many requests")
}

// ServiceUnavailable создает response с статусом 503 Service Unavailable
func ServiceUnavailable(c *gin.Context) {
	NewErrorResponse(c, http.StatusServiceUnavailable, "Service temporarily unavailable")
}

// ValidationError создает response с ошибками валидации
func ValidationError(c *gin.Context, validationErrors map[string]string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"error":   "Validation failed",
		"details": validationErrors,
	})
}
