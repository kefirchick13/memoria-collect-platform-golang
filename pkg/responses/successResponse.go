package responses

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type successResponse struct {
	Message string `json:"message"`
}

// NewSuccessResponse создает успешный JSON ответ
func NewSuccessResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, successResponse{Message: message})
}

// Success создает успешный ответ со статусом 200
func Success(c *gin.Context, message string) {
	c.JSON(http.StatusOK, successResponse{Message: message})
}

// Created создает ответ со статусом 201 (создано)
func Created(c *gin.Context, message string) {
	c.JSON(http.StatusCreated, successResponse{Message: message})
}
