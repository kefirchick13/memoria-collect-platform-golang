package service

import (
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/repository"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/pagination"
	"go.uber.org/zap"
)

type CollectionItemService struct {
	repo   repository.CollectionItem
	logger *zap.SugaredLogger
}

func NewCollectionItemService(repo repository.CollectionItem, logger *zap.SugaredLogger) *CollectionItemService {
	return &CollectionItemService{
		repo:   repo,
		logger: logger,
	}
}

func (s *CollectionItemService) GetItemsByCurrentType(currentType string, pagination pagination.PaginationRequest) (*pagination.PaginatedResponse[models.CollectionItem], error) {
	return s.repo.GetAllItemsWithCurrentTypePaginated(currentType, pagination)
}

func (s *CollectionItemService) GetItemsByCollection(collection_id string, user_id int) ([]models.CollectionItem, error) {
	return s.repo.GetItemsByCollection(collection_id)
}

func (s *CollectionItemService) GetItemByID(collection_item_id string) (*models.CollectionItem, error) {
	return s.repo.GetItemByID((collection_item_id))
}

func (s *CollectionItemService) CreateCollectionItem(item *models.CollectionItem) (string, error) {
	return s.repo.CreateItem(item)
}
func (s *CollectionItemService) DeleteItem(collection_item_id string) error {
	return s.repo.DeleteCollectionItem(collection_item_id)
}

func (s *CollectionItemService) UpdateItem(item *models.CollectionItem) error {
	return s.repo.UpdateCollectionItem(item)
}

func (s *CollectionItemService) AddItemToCollection(item_id string, collection_id string, user_review *string) (int, error) {
	if user_review == nil {
		return s.repo.AddItemToCollection(collection_id, item_id, "")
	} else {
		return s.repo.AddItemToCollection(collection_id, item_id, *user_review)
	}
}
