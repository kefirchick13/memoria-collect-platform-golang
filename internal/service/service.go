package service

import (
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/repository"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/pagination"
	"go.uber.org/zap"
)

type AuthService interface {
	Register(user *models.User) (*models.User, error)
	FindOrCreateGitHubUser(githubUser *models.GitHubUser) (*models.User, error)
	SignInWithPassword(mail string, password string) (models.User, string, error) // Возвращает токен
	SignInWithOAuth(user *models.User) (string, error)                            // Возвращает токен
	ParseToken(token string) (int, error)
}

type UserService interface {
	GetProfile(userID int) (*models.User, error)
	UpdateProfile(userID int, updates *UserProfileUpdate) error
	ChangePassword(userID int, currentPassword, newPassword string) error
	DeleteAccount(userID int) error
	LinkGitHubAccount(userID int, githubUser *models.GitHubUser) error
	UnlinkGitHubAccount(userID int, password string) error
	UpdateLastLogin(userID int) error
}

type CollectionService interface {
	CreateCollection(collection *models.Collection) (*models.Collection, error)
	GetCollections(user_id int) ([]models.Collection, error)
	GetCollectionsWithPagination(user_id int, pagination pagination.PaginationRequest) (*pagination.PaginatedResponse[models.Collection], error)
}

type CollectionItemService interface {
	GetItemsByCurrentType(currentType string, pagination pagination.PaginationRequest) (*pagination.PaginatedResponse[models.CollectionItem], error)
	GetItemsByCollection(collection_id string, user_id int) ([]models.CollectionItem, error)
	GetItemByID(collection_item_id string) (*models.CollectionItem, error)
	DeleteItem(collection_item_id string) error
	UpdateItem(item *models.CollectionItem) error
	AddItemToCollection(item_id string, collection_id string, user_review string, user_id int) (int, error)
	CreateCollectionItem(item *models.CollectionItem) (string, error)
}

type Service struct {
	AuthService
	UserService
	CollectionService
	CollectionItemService
}

func NewService(repository *repository.Repository, logger *zap.SugaredLogger) *Service {
	return &Service{
		AuthService:           NewAuthService(repository.UserRepository, logger),
		UserService:           NewUserService(repository.UserRepository, logger),
		CollectionService:     NewCollectionService(repository.Collection, logger),
		CollectionItemService: NewCollectionItemService(repository.CollectionItem, repository.Collection, logger),
	}
}
