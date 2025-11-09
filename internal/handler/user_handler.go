package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/service"
)

// UserProfileResponse represents user profile response
type UserProfileResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	GitHubID  *int64  `json:"github_id,omitempty"`
}

// UpdateUserProfileInput represents input for updating user profile
type UpdateUserProfileInput struct {
	Name      *string `json:"name,omitempty"`
	Email     *string `json:"email,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// ChangePasswordInput represents input for changing password
type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// LinkGitHubInput represents input for linking GitHub account
type LinkGitHubInput struct {
	GitHubCode string `json:"github_code" binding:"required"`
}

// UnlinkGitHubInput represents input for unlinking GitHub account
type UnlinkGitHubInput struct {
	Password string `json:"password" binding:"required"`
}

// MessageResponse represents simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// GetUserProfile получает профиль текущего пользователя
// @Summary Получить профиль пользователя
// @Description Возвращает профиль текущего аутентифицированного пользователя
// @Tags users
// @Produce json
// @Success 200 {object} UserProfileResponse "Профиль пользователя"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Security ApiKeyAuth
// @Router /users/profile [get]
func (h *Handler) GetUserProfile(c *gin.Context) {
	userID, _ := h.GetUserId(c)

	user, err := h.service.UserService.GetProfile(userID)
	if err != nil {
		h.logger.Errorf("Failed to get user profile for ID %d: %v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// UpdateUserProfile обновляет профиль пользователя
// @Summary Обновить профиль пользователя
// @Description Обновляет данные профиля текущего пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param input body UpdateUserProfileInput true "Данные для обновления профиля"
// @Success 200 {object} MessageResponse "Профиль успешно обновлен"
// @Failure 400 {object} ErrorResponse "Неверные входные данные"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 409 {object} ErrorResponse "Email уже используется"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /users/profile [put]
func (h *Handler) UpdateUserProfile(c *gin.Context) {
	userID, _ := h.GetUserId(c)

	var input UpdateUserProfileInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидация входных данных
	if input.Name != nil && len(*input.Name) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name must be at least 2 characters long"})
		return
	}

	if input.Email != nil && !isValidEmail(*input.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}

	updates := &service.UserProfileUpdate{
		Name:      input.Name,
		Email:     input.Email,
		AvatarURL: input.AvatarURL,
	}

	if err := h.service.UserService.UpdateProfile(userID, updates); err != nil {
		h.logger.Errorf("Failed to update profile for user %d: %v", userID, err)

		switch err.Error() {
		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case "email already in use":
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		}
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "profile updated successfully"})
}

// ChangePassword изменяет пароль пользователя
// @Summary Изменить пароль
// @Description Изменяет пароль текущего пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param input body ChangePasswordInput true "Данные для изменения пароля"
// @Success 200 {object} MessageResponse "Пароль успешно изменен"
// @Failure 400 {object} ErrorResponse "Неверные входные данные"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /users/password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, _ := h.GetUserId(c)

	var input ChangePasswordInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UserService.ChangePassword(userID, input.CurrentPassword, input.NewPassword); err != nil {
		h.logger.Errorf("Failed to change password for user %d: %v", userID, err)

		switch err.Error() {
		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case "password not set for this account":
			c.JSON(http.StatusBadRequest, gin.H{"error": "password not set for this account"})
		case "current password is incorrect":
			c.JSON(http.StatusBadRequest, gin.H{"error": "current password is incorrect"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to change password"})
		}
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "password changed successfully"})
}

// DeleteAccount удаляет аккаунт пользователя
// @Summary Удалить аккаунт
// @Description Полностью удаляет аккаунт текущего пользователя
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse "Аккаунт успешно удален"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /users/account [delete]
func (h *Handler) DeleteAccount(c *gin.Context) {
	userID, _ := h.GetUserId(c)

	if err := h.service.UserService.DeleteAccount(userID); err != nil {
		h.logger.Errorf("Failed to delete account for user %d: %v", userID, err)

		switch err.Error() {
		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case "invalid password":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete account"})
		}
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "account deleted successfully"})
}

// LinkGitHubAccount привязывает GitHub аккаунт к существующему пользователю
// @Summary Привязать GitHub аккаунт
// @Description Привязывает GitHub аккаунт к текущему пользователю
// @Tags users
// @Accept json
// @Produce json
// @Param input body LinkGitHubInput true "Код авторизации GitHub"
// @Success 200 {object} MessageResponse "GitHub аккаунт успешно привязан"
// @Failure 400 {object} ErrorResponse "Неверные входные данные"
// @Failure 409 {object} ErrorResponse "GitHub аккаунт уже привязан к другому пользователю"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /users/github/link [post]
func (h *Handler) LinkGitHubAccount(c *gin.Context) {
	userID, _ := h.GetUserId(c)

	var input LinkGitHubInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем GitHub пользователя по коду
	githubAccessToken, err := h.exchangeCodeForToken(input.GitHubCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to exchange GitHub code"})
		return
	}

	githubUser, err := h.getGitHubUser(githubAccessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get GitHub user data"})
		return
	}

	if err := h.service.UserService.LinkGitHubAccount(userID, githubUser); err != nil {
		h.logger.Errorf("Failed to link GitHub account for user %d: %v", userID, err)

		switch err.Error() {
		case "GitHub account already linked to another user":
			c.JSON(http.StatusConflict, gin.H{"error": "GitHub account already linked to another user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link GitHub account"})
		}
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "GitHub account linked successfully"})
}

// UnlinkGitHubAccount отвязывает GitHub аккаунт от пользователя
// @Summary Отвязать GitHub аккаунт
// @Description Отвязывает GitHub аккаунт от текущего пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param input body UnlinkGitHubInput true "Пароль для подтверждения"
// @Success 200 {object} MessageResponse "GitHub аккаунт успешно отвязан"
// @Failure 400 {object} ErrorResponse "Неверные входные данные"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /users/github/unlink [post]
func (h *Handler) UnlinkGitHubAccount(c *gin.Context) {
	userID, _ := h.GetUserId(c)

	var input UnlinkGitHubInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UserService.UnlinkGitHubAccount(userID, input.Password); err != nil {
		h.logger.Errorf("Failed to unlink GitHub account for user %d: %v", userID, err)

		switch err.Error() {
		case "cannot unlink GitHub account without setting a password first":
			c.JSON(http.StatusBadRequest, gin.H{"error": "set a password first before unlinking GitHub"})
		case "invalid password":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlink GitHub account"})
		}
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "GitHub account unlinked successfully"})
}

// isValidEmail проверяет валидность email (базовая проверка)
func isValidEmail(email string) bool {
	if !contains(email, "@") {
		return false
	}
	if !contains(email, ".") {
		return false
	}
	return true
}

// contains проверяет наличие подстроки в строке
func contains(str, subStr string) bool {
	if len(subStr) > len(str) {
		return false
	}
	if len(subStr) == 0 {
		return true
	}
	if str == subStr {
		return true
	}

	for i := 0; i <= len(str)-len(subStr); i++ {
		if str[i:i+len(subStr)] == subStr {
			return true
		}
	}

	return false
}
