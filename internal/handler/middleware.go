package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/pagination"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "userId"
)

func (h *Handler) userIdentity(c *gin.Context) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "errorAuthHeader",
		})
		c.Abort()
		return
	}
	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid auth header",
		})
		c.Abort()
		return
	}

	userId, err := h.service.AuthService.ParseToken(headerParts[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": userId,
		})
		c.Abort()
		return
	}

	c.Set(userCtx, userId)

}

func (h *Handler) GetUserId(c *gin.Context) (int, error) {
	id, ok := c.Get(userCtx)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user is not found",
		})
		c.Abort()
		return 0, errors.New("error id is not found")
	}
	idInt, ok := id.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "userId is not valid type",
		})
		c.Abort()
		return 0, errors.New("error id is not found")
	}
	return idInt, nil

}

func GetPaginationParams(c *gin.Context) pagination.PaginationRequest {
	// Парсим строки в числа
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return pagination.DefaultPaginationRequest()
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return pagination.DefaultPaginationRequest()
	}

	return pagination.NewPaginationRequest(page, limit)
}
