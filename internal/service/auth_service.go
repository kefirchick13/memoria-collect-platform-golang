package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	tokenExpiredTime = time.Hour * 24 * 7
	tokenSignInKey   = "orfdsjl43785652312089" // Используйте одну константу для подписи!
)

type AuthService struct {
	repo   repository.Authorization
	logger *zap.SugaredLogger
}

func NewAuthService(repo repository.Authorization, logger *zap.SugaredLogger) *AuthService {
	return &AuthService{
		repo:   repo,
		logger: logger,
	}
}

func (s *AuthService) CreateUser(user *models.User) (*models.User, error) {
	// Хэшируем пароль только если он предоставлен (для email регистрации)
	if user.Password != nil {
		hashedPassword, err := s.hashPassword(*user.Password)
		if err != nil {
			return nil, err
		}
		user.Password = &hashedPassword
	}

	return s.repo.CreateUser(user)
}

// Новый метод для поиска/создания GitHub пользователя
func (s *AuthService) FindOrCreateGitHubUser(githubUser *models.GitHubUser) (*models.User, error) {
	// 1. Ищем по GitHub ID
	existingUser, err := s.repo.GetUserByGitHubID(githubUser.ID)
	if err == nil && existingUser != nil {
		return existingUser, nil
	}

	// 2. Если не нашли, ищем по email
	if githubUser.Email != nil && *githubUser.Email != "" {
		existingUser, err = s.repo.GetUserByEmail(*githubUser.Email)

		if err == nil && existingUser != nil {
			// Привязываем GitHub к существующему аккаунту
			return s.repo.LinkGitHubToExistingUser(existingUser.ID, githubUser)
		}
	}

	// 3. Создаем нового пользователя
	newUser := &models.User{
		Name:        *githubUser.Name,
		Mail:        *githubUser.Email,
		GitHubID:    &githubUser.ID,
		GitHubLogin: &githubUser.Login,
		AvatarURL:   &githubUser.AvatarURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	jsonUser, err := json.Marshal(newUser)
	if err != nil {
		s.logger.Error(err.Error())
	} else {
		s.logger.Infof("json user: %s", jsonUser)
	}

	return s.repo.CreateUser(newUser)
}

// Для email аутентификации (проверяет пароль)
func (s *AuthService) SignInWithPassword(mail string, password string) (string, error) {
	// 1. Ищем пользователя по email
	user, err := s.repo.GetUserByEmail(mail)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}

	// 3. Проверяем пароль
	if user.Password == nil {
		return "", fmt.Errorf("invalid password")
	}

	if !s.checkPasswordHash(password, *user.Password) {
		return "", fmt.Errorf("invalid password")
	}

	// 4. Генерируем JWT токен
	return s.generateJWTToken(user.ID, user.Name)
}

// Для GitHub аутентификации (без проверки пароля)
func (s *AuthService) SignInWithOAuth(user *models.User) (string, error) {
	return s.generateJWTToken(user.ID, user.Name)
}

type CustomClaims struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

func (s *AuthService) generateJWTToken(userID int, name string) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiredTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSignInKey))
}

func (s *AuthService) ParseToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSignInKey), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	return claims.UserID, nil
}

// Методы для работы с паролями
func (s *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *AuthService) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
