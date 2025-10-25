package service

import (
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/repository"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/pagination"
	"go.uber.org/zap"
)

type Authorization interface {
	CreateUser(user *models.User) (*models.User, error)
	FindOrCreateGitHubUser(githubUser *models.GitHubUser) (*models.User, error)
	SignInWithPassword(mail string, password string) (string, error) // Возвращает токен
	SignInWithOAuth(user *models.User) (string, error)               // Возвращает токен
	ParseToken(token string) (int, error)
}

type Collection interface {
	CreateCollection(collection *models.Collection) (*models.Collection, error)
	GetCollections(user_id int) ([]models.Collection, error)
	GetCollectionsWithPagination(user_id int, pagination pagination.PaginationRequest) (*pagination.PaginatedResponse[models.Collection], error)
}

type CollectionItems interface {
	GetItemsByCurrentType(currentType string, pagination pagination.PaginationRequest) (*pagination.PaginatedResponse[models.CollectionItem], error)
	GetItemsByCollection(collection_id string, user_id int) ([]models.CollectionItem, error)
	GetItemByID(collection_item_id string) (*models.CollectionItem, error)
	DeleteItem(collection_item_id string) error
	UpdateItem(item *models.CollectionItem) error
	AddItemToCollection(item_id string, collection_id string, user_review *string) (int, error)
	CreateCollectionItem(item *models.CollectionItem) (string, error)
}

type Service struct {
	Authorization
	Collection
	CollectionItems
}

func NewService(repository *repository.Repository, logger *zap.SugaredLogger) *Service {
	return &Service{
		Authorization:   NewAuthService(repository.Authorization, logger),
		Collection:      NewCollectionService(repository.Collection, logger),
		CollectionItems: NewCollectionItemService(repository.CollectionItem, logger),
	}
}
