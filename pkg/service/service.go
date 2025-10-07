package service

import (
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/repository"
	"go.uber.org/zap"
)

type Authorization interface {
	CreateUser(user *models.User) (*models.User, error)
	FindOrCreateGitHubUser(githubUser *models.GitHubUser) (*models.User, error)
	SignInWithPassword(mail string, password string) (string, error) // Возвращает токен
	SignInWithOAuth(user *models.User) (string, error)               // Возвращает токен
	ParseToken(token string) (string, error)
}

type Service struct {
	logger *zap.SugaredLogger
	Authorization
}

func NewService(repository *repository.Repository, logger *zap.SugaredLogger) *Service {
	return &Service{
		logger:        logger,
		Authorization: NewAuthService(repository.Authorization, logger),
	}
}
