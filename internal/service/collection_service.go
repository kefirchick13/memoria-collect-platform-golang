package service

import (
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/repository"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/pagination"
	"go.uber.org/zap"
)

type collectionService struct {
	repo   repository.Collection
	logger *zap.SugaredLogger
}

func NewCollectionService(repo repository.Collection, logger *zap.SugaredLogger) *collectionService {
	return &collectionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *collectionService) CreateCollection(collection *models.Collection) (*models.Collection, error) {
	return s.repo.CreateCollection(collection)
}

func (s *collectionService) GetCollections(user_id int) ([]models.Collection, error) {
	return s.repo.GetCollections(user_id)
}

func (s *collectionService) GetCollectionsWithPagination(user_id int, pagination pagination.PaginationRequest) (*pagination.PaginatedResponse[models.Collection], error) {
	return s.repo.GetCollectionsWithPagination(user_id, pagination)
}
