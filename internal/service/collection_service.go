package service

import (
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/repository"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/pagination"
	"go.uber.org/zap"
)

type CollectionService struct {
	repo   repository.Collection
	logger *zap.SugaredLogger
}

func NewCollectionService(repo repository.Collection, logger *zap.SugaredLogger) *CollectionService {
	return &CollectionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *CollectionService) CreateCollection(collection *models.Collection) (*models.Collection, error) {
	return s.repo.CreateCollection(collection)
}

func (s *CollectionService) GetCollections(user_id int) ([]models.Collection, error) {
	return s.repo.GetCollections(user_id)
}

func (s *CollectionService) GetCollectionsWithPagination(user_id int, pagination pagination.PaginationRequest) (*pagination.PaginatedResponse[models.Collection], error) {
	return s.repo.GetCollectionsWithPagination(user_id, pagination)
}
