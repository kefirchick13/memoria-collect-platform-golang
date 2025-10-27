package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/service"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/responses"
)

func (h *Handler) GetItemsByType(c *gin.Context) {
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
	user_id, err := GetUserId(c)
	if err != nil {
		return
	}
	collection_uid := c.Param("id")
	if collection_uid == "" {
		responses.NewErrorResponse(c, http.StatusBadRequest, "id hasn't been provided")
		return
	}

	items, err := h.service.CollectionItems.GetItemsByCollection(collection_uid, user_id)
	if err != nil {
		responses.NewErrorResponse(c, http.StatusNotFound, "items aren't find")
		return
	}
	c.JSON(http.StatusOK, items)
}

type AddItemInput struct {
	UserReview string `json:"user_review"`
	ItemId     string `json:"item_id"`
}

func (h *Handler) AddItemToCollection(c *gin.Context) {
	user_id, err := GetUserId(c)
	if err != nil {
		return
	}
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

	_, err = h.service.CollectionItems.AddItemToCollection(input.ItemId, collection_uid, input.UserReview, user_id)
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

type CreateCollectionInput struct {
	Type        string  `json:"type" db:"type"`
	Title       string  `json:"title" db:"title"`
	Description string  `json:"description" db:"description"`
	CoverImage  *string `json:"cover_image" db:"cover_image"`
	IsCustom    bool    `json:"is_custom" db:"is_custom"`
	IsPublic    bool    `json:"is_public" db:"is_public"`
}

func (h *Handler) CreateCollectionItem(c *gin.Context) {
	user_id, err := GetUserId(c)
	if err != nil {
		return
	}
	var input CreateCollectionInput

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
