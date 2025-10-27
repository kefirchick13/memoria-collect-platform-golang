package repository

import (
	"database/sql"

	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/pagination"
	"go.uber.org/zap"
)

type Authorization interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)    // Основной метод для поиска
	GetUserByGitHubID(githubID int) (*models.User, error) // Для GitHub аутентификации
	LinkGitHubToExistingUser(userID int, githubUser *models.GitHubUser) (*models.User, error)
}

type Collection interface {
	CreateCollection(collection *models.Collection) (*models.Collection, error)
	GetCollections(user_id int) ([]models.Collection, error)
	DeleteCollection(collectionID string) error
	GetCollectionByID(collectionID string) (*models.Collection, error)
	GetCollectionsWithPagination(userID int, req pagination.PaginationRequest) (*pagination.PaginatedResponse[models.Collection], error)
}

type CollectionItem interface {
	GetAllItemsWithCurrentTypePaginated(currentType string, pagination pagination.PaginationRequest) (*pagination.PaginatedResponse[models.CollectionItem], error)
	GetItemsByCollection(collection_id string) ([]models.CollectionItem, error)
	GetItemByID(id string) (*models.CollectionItem, error)
	CreateItem(collectionItem *models.CollectionItem) (string, error)
	DeleteCollectionItem(id string) error
	UpdateCollectionItem(item *models.CollectionItem) error
	AddItemToCollection(collection_id string, item_id string, user_review string) (int, error)
}

type Repository struct {
	Authorization
	Collection
	CollectionItem
}

func NewRepository(db *sql.DB, logger *zap.SugaredLogger) *Repository {
	return &Repository{
		Authorization:  NewAuthPostgres(db, logger),
		Collection:     NewCollectionPostgres(db, logger),
		CollectionItem: NewCollectionItemPostgres(db, logger),
	}
}
