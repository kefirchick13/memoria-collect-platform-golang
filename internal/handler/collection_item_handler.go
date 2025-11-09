package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/service"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/responses"
)

// CollectionItemResponse represents a collection item response
type CollectionItemResponse struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	CoverImage  *string `json:"cover_image"`
	IsCustom    bool    `json:"is_custom"`
	IsPublic    bool    `json:"is_public"`
	CreatorID   *string `json:"creator_id"`
}

// PaginatedItemsResponse represents paginated items response
type PaginatedItemsResponse struct {
	Items      []CollectionItemResponse `json:"items"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	PerPage    int                      `json:"per_page"`
	TotalPages int                      `json:"total_pages"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// IDResponse represents ID response
type IDResponse struct {
	ID string `json:"id"`
}

// GetItemsByType returns paginated list of items by type
// @Summary Получить элементы по типу
// @Description Возвращает пагинированный список элементов по указанному типу
// @Tags items
// @Produce json
// @Param type query string false "Тип элементов" default(book)
// @Param page query int false "Номер страницы" default(1)
// @Param per_page query int false "Кол-во элементов на странице" default(10)
// @Success 200 {object} PaginatedItemsResponse "Успешный ответ"
// @Failure 502 {object} ErrorResponse "Ошибка сервера"
// @Router /items [get]
func (h *Handler) GetItemsByType(c *gin.Context) {
	// search for a type in query params, Default as "book"
	currType := c.DefaultQuery("type", "book")
	pagination := GetPaginationParams(c)
	PaginatedResponse, err := h.service.CollectionItemService.GetItemsByCurrentType(currType, pagination)

	if err != nil {
		h.logger.Errorf("Error during getting items by type: %s", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PaginatedResponse)
}

// GetCollectionItems returns all items in a collection
// @Summary Получить элементы коллекции
// @Description Возвращает все элементы указанной коллекции
// @Tags collections
// @Produce json
// @Param id path string true "ID коллекции"
// @Success 200 {object} interface{} "Список элементов коллекции"
// @Failure 400 {object} ErrorResponse "Не указан ID коллекции"
// @Failure 404 {object} ErrorResponse "Элементы не найдены"
// @Security ApiKeyAuth
// @Router /collections/{id}/items [get]
func (h *Handler) GetCollectionItems(c *gin.Context) {
	user_id, _ := h.GetUserId(c)

	collection_uid := c.Param("id")
	if collection_uid == "" {
		responses.NewErrorResponse(c, http.StatusBadRequest, "id hasn't been provided")
		return
	}

	items, err := h.service.CollectionItemService.GetItemsByCollection(collection_uid, user_id)
	if err != nil {
		responses.NewErrorResponse(c, http.StatusNotFound, "items aren't find")
		return
	}
	c.JSON(http.StatusOK, items)
}

// AddItemInput represents input for adding item to collection
type AddItemInput struct {
	UserReview string `json:"user_review" binding:"required"`
	ItemId     string `json:"item_id" binding:"required"`
}

// AddItemToCollection adds an item to a collection
// @Summary Добавить элемент в коллекцию
// @Description Добавляет существующий элемент в указанную коллекцию
// @Tags collections
// @Accept json
// @Produce json
// @Param id path string true "ID коллекции"
// @Param input body AddItemInput true "Данные для добавления элемента"
// @Success 200 {object} SuccessResponse "Элемент успешно добавлен"
// @Failure 400 {object} ErrorResponse "Неверные входные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещен"
// @Failure 404 {object} ErrorResponse "Коллекция или элемент не найдены"
// @Failure 409 {object} ErrorResponse "Элемент уже в коллекции"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /collections/{id}/items [post]
func (h *Handler) AddItemToCollection(c *gin.Context) {
	user_id, _ := h.GetUserId(c)

	collection_uid := c.Param("id")
	if collection_uid == "" {
		responses.NewErrorResponse(c, http.StatusBadRequest, "id hasn't been provided")
		return
	}

	var input AddItemInput

	if err := c.BindJSON(&input); err != nil {
		responses.NewErrorResponse(c, http.StatusBadRequest, "input is not valid")
		return
	}

	_, err := h.service.CollectionItemService.AddItemToCollection(input.ItemId, collection_uid, input.UserReview, user_id)
	if err != nil {
		switch err {
		case service.ErrCollectionNotFound:
			responses.NewErrorResponse(c, http.StatusNotFound, "Collection not found")
		case service.ErrItemNotFound:
			responses.NewErrorResponse(c, http.StatusNotFound, "Item not found")
		case service.ErrNotCollectionOwner:
			responses.NewErrorResponse(c, http.StatusForbidden, "Access denied")
		case service.ErrItemAlreadyInCollection:
			responses.NewErrorResponse(c, http.StatusConflict, "Item already in collection")
		default:
			responses.NewErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to add item to collection err: %s", err.Error()))
		}
		return
	}

	responses.Success(c, "Item had assigned to the list")
}

// CreateCollectionItemInput represents input for creating a collection item
type CreateCollectionItemInput struct {
	Type        string  `json:"type" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	CoverImage  *string `json:"cover_image"`
	IsCustom    bool    `json:"is_custom"`
	IsPublic    bool    `json:"is_public"`
}

// CreateCollectionItem creates a new collection item
// @Summary Создать элемент коллекции
// @Description Создает новый элемент коллекции
// @Tags items
// @Accept json
// @Produce json
// @Param input body CreateCollectionItemInput true "Данные для создания элемента"
// @Success 200 {object} IDResponse "Элемент создан, возвращает ID"
// @Failure 400 {object} ErrorResponse "Неверные входные данные"
// @Security ApiKeyAuth
// @Router /items [post]
func (h *Handler) CreateCollectionItem(c *gin.Context) {
	user_id, _ := h.GetUserId(c)

	var input CreateCollectionItemInput

	if err := c.BindJSON(&input); err != nil {
		responses.NewErrorResponse(c, http.StatusBadRequest, "input is not valid")
		return
	}

	collection_item := models.CollectionItem{
		Type:        input.Type,
		Title:       input.Title,
		Description: input.Description,
		CoverImage:  input.CoverImage,
		IsCustom:    input.IsCustom,
		IsPublic:    input.IsPublic,
		CreatorID:   &user_id,
	}
	id, err := h.service.CollectionItemService.CreateCollectionItem(&collection_item)

	if err != nil {
		h.logger.Errorf("Can't create collectionsItem, error: %d", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}
