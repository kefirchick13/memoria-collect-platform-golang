package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
)

func (h *Handler) GetCollectionList(c *gin.Context) {
	userId, err := GetUserId(c)
	pagination := GetPaginationParams(c)

	if err != nil {
		return
	}

	PaginatedResponse, err := h.service.Collection.GetCollectionsWithPagination(userId, pagination)
	if err != nil {
		h.logger.Errorf("Failed to get collections for user %d: %v", userId, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to recieved collections",
		})
		return
	}

	c.JSON(http.StatusOK, PaginatedResponse)
}

func (h *Handler) CreateCollection(c *gin.Context) {
	user_id, err := GetUserId(c)
	h.logger.Info("user_id: ", user_id)
	if err != nil {
		return
	}
	var input struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		IsPublic    bool    `json:"is_public"`
		IsCustom    bool    `json:"is_custom"`
		CoverImage  *string `json:"cover_image"`
		Type        string  `json:"type" binding:"required"`
	}

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

	newCollection, err := h.service.Collection.CreateCollection(&collection)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusCreated, newCollection)

}
