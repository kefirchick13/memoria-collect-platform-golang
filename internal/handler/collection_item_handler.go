package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
)

func (h *Handler) GetCollectionItemsByType(c *gin.Context) {
	// search for a type in query params, Default as "book"
	currType := c.DefaultQuery("type", "book")
	pagination := GetPaginationParams(c)
	PaginatedResponse, err := h.service.CollectionItems.GetItemsByCurrentType(currType, pagination)

	if err != nil {
		h.logger.Errorf("Error during getting items by type: %s", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusOK, PaginatedResponse)
}

func (h *Handler) GetCollectionItems(c *gin.Context) {

}

func (h *Handler) CreateCollectionItem(c *gin.Context) {
	user_id, err := GetUserId(c)
	if err != nil {
		return
	}
	var input struct {
		Type        string  `json:"type" db:"type"`
		Title       string  `json:"title" db:"title"`
		Description string  `json:"description" db:"description"`
		CoverImage  *string `json:"cover_image" db:"cover_image"`
		IsCustom    bool    `json:"is_custom" db:"is_custom"`
		IsPublic    bool    `json:"is_public" db:"is_public"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	id, err := h.service.CollectionItems.CreateCollectionItem(&collection_item)

	if err != nil {
		h.logger.Errorf("Cant create collectionsItem, error: %d", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}
