package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
)

// CollectionResponse represents a collection response
type CollectionResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	IsPublic    bool    `json:"is_public"`
	IsCustom    bool    `json:"is_custom"`
	CoverImage  *string `json:"cover_image"`
	Type        string  `json:"type"`
	UserID      string  `json:"user_id"`
}

// PaginatedCollectionsResponse represents paginated collections response
type PaginatedCollectionsResponse struct {
	Collections []CollectionResponse `json:"collections"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	PerPage     int                  `json:"per_page"`
	TotalPages  int                  `json:"total_pages"`
}

// CreateCollectionInput represents input for creating a collection
type CreateCollectionInput struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	IsPublic    bool    `json:"is_public"`
	IsCustom    bool    `json:"is_custom"`
	CoverImage  *string `json:"cover_image"`
	Type        string  `json:"type" binding:"required"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// GetCollectionList returns paginated list of collections for user
// @Summary Получить список коллекций
// @Description Возвращает пагинированный список коллекций пользователя
// @Tags collections
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param per_page query int false "Кол-во элементов на странице" default(10)
// @Success 200 {object} PaginatedCollectionsResponse "Успешный ответ"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /collections [get]
func (h *Handler) GetCollectionList(c *gin.Context) {
	userId, _ := h.GetUserId(c)
	pagination := GetPaginationParams(c)

	PaginatedResponse, err := h.service.CollectionService.GetCollectionsWithPagination(userId, pagination)
	if err != nil {
		h.logger.Errorf("Failed to get collections for user %d: %v", userId, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to recieved collections",
		})
		return
	}

	c.JSON(http.StatusOK, PaginatedResponse)
}

// CreateCollection creates a new collection
// @Summary Создать коллекцию
// @Description Создает новую коллекцию для пользователя
// @Tags collections
// @Accept json
// @Produce json
// @Param input body CreateCollectionInput true "Данные для создания коллекции"
// @Success 201 {object} CollectionResponse "Коллекция создана"
// @Failure 400 {object} ErrorResponse "Ошибка валидации"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /collections [post]
func (h *Handler) CreateCollection(c *gin.Context) {
	user_id, _ := h.GetUserId(c)

	var input CreateCollectionInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := models.Collection{
		Name:        input.Name,
		Description: input.Description,
		IsPublic:    input.IsPublic,
		CoverImage:  input.CoverImage,
		Type:        input.Type,
		UserID:      user_id,
	}

	newCollection, err := h.service.CollectionService.CreateCollection(&collection)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusCreated, newCollection)
}
