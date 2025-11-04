package service

import (
	"fmt"
	"time"

	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/repository"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/password_hash"
	"go.uber.org/zap"
)

type userService struct {
	repo   repository.UserRepository
	logger *zap.SugaredLogger
}

func NewUserService(repo repository.UserRepository, logger *zap.SugaredLogger) *userService {
	return &userService{
		repo:   repo,
		logger: logger,
	}
}

type UserProfileUpdate struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
}

func (s *userService) GetProfile(userID int) (*models.User, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		s.logger.Errorf("Failed to get user profile for ID %d: %v", userID, err)
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *userService) UpdateProfile(userID int, updates *UserProfileUpdate) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	currentUser, err := s.repo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Применяем обновления
	if updates.Name != nil {
		currentUser.Name = *updates.Name
	}
	if updates.Email != nil {
		// Проверяем, не используется ли email другим пользователем
		existingUser, err := s.repo.GetUserByEmail(*updates.Email)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return fmt.Errorf("email already in use")
		}
		currentUser.Mail = *updates.Email
	}
	if updates.AvatarURL != nil {
		currentUser.AvatarURL = updates.AvatarURL
	}

	currentUser.UpdatedAt = time.Now()

	if err := s.repo.UpdateUser(currentUser); err != nil {
		s.logger.Errorf("Failed to update user profile for ID %d: %v", userID, err)
		return fmt.Errorf("failed to update profile")
	}

	return nil
}

func (s *userService) ChangePassword(userID int, currentPassword, newPassword string) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	if len(newPassword) < 8 {
		return fmt.Errorf("new password must be at least 8 characters long")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Проверяем текущий пароль
	if user.Password == nil {
		return fmt.Errorf("password not set for this account")
	}

	if !password_hash.Check(currentPassword, *user.Password) {
		return fmt.Errorf("current password is incorrect")
	}

	// Хешируем новый пароль
	hashedPassword, err := password_hash.Hash(newPassword)
	if err != nil {
		s.logger.Errorf("Failed to hash password for user %d: %v", userID, err)
		return fmt.Errorf("failed to change password")
	}

	// Обновляем пароль
	if err := s.repo.UpdateUserPassword(userID, hashedPassword); err != nil {
		s.logger.Errorf("Failed to update password for user %d: %v", userID, err)
		return fmt.Errorf("failed to change password")
	}

	return nil
}

func (s *userService) DeleteAccount(userID int) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	// Выполняем soft delete
	if err := s.repo.DeleteUser(userID); err != nil {
		s.logger.Errorf("Failed to delete user account %d: %v", userID, err)
		return fmt.Errorf("failed to delete account")
	}

	return nil
}

func (s *userService) LinkGitHubAccount(userID int, githubUser *models.GitHubUser) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	// Проверяем, не привязан ли GitHub аккаунт к другому пользователю
	existingUser, err := s.repo.GetUserByGitHubID(githubUser.ID)
	if err == nil && existingUser != nil && existingUser.ID != userID {
		return fmt.Errorf("GitHub account already linked to another user")
	}

	// Привязываем GitHub аккаунт
	_, err = s.repo.LinkGitHubToExistingUser(userID, githubUser)
	if err != nil {
		s.logger.Errorf("Failed to link GitHub account for user %d: %v", userID, err)
		return fmt.Errorf("failed to link GitHub account")
	}

	return nil
}

func (s *userService) UnlinkGitHubAccount(userID int, password string) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Проверяем, что у пользователя есть пароль (чтобы не заблокировать вход)
	if user.Password == nil {
		return fmt.Errorf("cannot unlink GitHub account without setting a password first")
	}

	// Проверяем пароль
	if !password_hash.Check(password, *user.Password) {
		return fmt.Errorf("invalid password")
	}

	// Отвязываем GitHub аккаунт
	user.GitHubID = nil
	user.GitHubLogin = nil
	user.UpdatedAt = time.Now()

	if err := s.repo.UpdateUser(user); err != nil {
		s.logger.Errorf("Failed to unlink GitHub account for user %d: %v", userID, err)
		return fmt.Errorf("failed to unlink GitHub account")
	}

	return nil
}

func (s *userService) UpdateLastLogin(userID int) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	if err := s.repo.UpdateLastLogin(userID); err != nil {
		s.logger.Warnf("Failed to update last login for user %d: %v", userID, err)
		return err
	}

	return nil
}
