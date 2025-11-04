package service

import (
	"errors"

	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/repository"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/pagination"
	"go.uber.org/zap"
)

var (
	ErrCollectionNotFound      = errors.New("collection not found")
	ErrItemNotFound            = errors.New("item not found")
	ErrNotCollectionOwner      = errors.New("user is not collection owner")
	ErrItemAlreadyInCollection = errors.New("item already exists in collection")
)

type collectionItemService struct {
	itemRepo       repository.CollectionItem
	collectionRepo repository.Collection
	logger         *zap.SugaredLogger
}

func NewCollectionItemService(itemRepo repository.CollectionItem, collectionRepo repository.Collection, logger *zap.SugaredLogger) *collectionItemService {
	return &collectionItemService{
		itemRepo:       itemRepo,
		collectionRepo: collectionRepo,
		logger:         logger,
	}
}

func (s *collectionItemService) GetItemsByCurrentType(currentType string, pagination pagination.PaginationRequest) (*pagination.PaginatedResponse[models.CollectionItem], error) {
	return s.itemRepo.GetAllItemsWithCurrentTypePaginated(currentType, pagination)
}

func (s *collectionItemService) GetItemsByCollection(collection_id string, user_id int) ([]models.CollectionItem, error) {
	return s.itemRepo.GetItemsByCollection(collection_id)
}

func (s *collectionItemService) GetItemByID(collection_item_id string) (*models.CollectionItem, error) {
	return s.itemRepo.GetItemByID((collection_item_id))
}

func (s *collectionItemService) CreateCollectionItem(item *models.CollectionItem) (string, error) {
	return s.itemRepo.CreateItem(item)
}
func (s *collectionItemService) DeleteItem(collection_item_id string) error {
	return s.itemRepo.DeleteCollectionItem(collection_item_id)
}

func (s *collectionItemService) UpdateItem(item *models.CollectionItem) error {
	return s.itemRepo.UpdateCollectionItem(item)
}

func (s *collectionItemService) AddItemToCollection(item_id string, collection_id string, user_review string, user_id int) (int, error) {
	// Проверяем существование коллекции и права доступа
	collection, err := s.collectionRepo.GetCollectionByID(collection_id)
	if err != nil {
		return 0, ErrCollectionNotFound
	}
	if collection.UserID != user_id {
		return 0, ErrNotCollectionOwner
	}

	// Проверяем существование item
	item, err := s.itemRepo.GetItemByID(item_id)
	if err != nil {
		return 0, ErrItemNotFound
	}

	// Проверяем, что item подходит по типу к коллекции
	if item.Type != collection.Type {
		return 0, errors.New("item type doesn't match collection type")
	}

	// Проверяем, что item публичный или создан этим пользователем
	if !item.IsPublic && (item.CreatorID == nil || *item.CreatorID != user_id) {
		return 0, errors.New("access denied to this item")
	}

	return s.itemRepo.AddItemToCollection(collection_id, item_id, user_review)
}
