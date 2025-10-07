package repository

import (
	"database/sql"

	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/models"
	"go.uber.org/zap"
)

type Authorization interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)    // Основной метод для поиска
	GetUserByGitHubID(githubID int) (*models.User, error) // Для GitHub аутентификации
	LinkGitHubToExistingUser(userID int, githubUser *models.GitHubUser) (*models.User, error)
}

type Repository struct {
	Authorization
}

func NewRepository(db *sql.DB, logger *zap.SugaredLogger) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db, logger),
	}
}
