package responses

import (
	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(c *gin.Context, statusCode int, err string) {
	c.AbortWithStatusJSON(statusCode, errorResponse{Error: err})
}
