package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/service"
)

// GetUserProfile получает профиль текущего пользователя
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
func (h *Handler) UpdateUserProfile(c *gin.Context) {
	userID, _ := h.GetUserId(c)

	var input struct {
		Name      *string `json:"name,omitempty"`
		Email     *string `json:"email,omitempty"`
		AvatarURL *string `json:"avatar_url,omitempty"`
	}

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

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully"})
}

// ChangePassword изменяет пароль пользователя
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, _ := h.GetUserId(c)

	var input struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

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

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// DeleteAccount удаляет аккаунт пользователя
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

	c.JSON(http.StatusOK, gin.H{"message": "account deleted successfully"})
}

// LinkGitHubAccount привязывает GitHub аккаунт к существующему пользователю
func (h *Handler) LinkGitHubAccount(c *gin.Context) {
	userID, _ := h.GetUserId(c)

	var input struct {
		GitHubCode string `json:"github_code" binding:"required"`
	}

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

	c.JSON(http.StatusOK, gin.H{"message": "GitHub account linked successfully"})
}

// UnlinkGitHubAccount отвязывает GitHub аккаунт от пользователя
func (h *Handler) UnlinkGitHubAccount(c *gin.Context) {
	userID, _ := h.GetUserId(c)

	var input struct {
		Password string `json:"password" binding:"required"`
	}

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

	c.JSON(http.StatusOK, gin.H{"message": "GitHub account unlinked successfully"})
}

func isValidEmail(email string) bool {
	if !contains(email, "@") {
		return false
	}
	if !contains(email, ".") {
		return false
	}
	return true
}

func contains(str, subStr string) bool {
	// Если подстрока больше - значит в строке точно не может быть подстроки
	if len(subStr) > len(str) {
		return false
	}
	// Если подстрока пустая - считается что содержится в любой строке
	if len(subStr) == 0 {
		return true
	}

	// Если строки идентичны - сразу возвращаем true
	if str == subStr {
		return true
	}

	for i := 0; i <= len(str)-len(subStr); i++ {
		if str[:i+len(subStr)] == subStr {
			return true
		}
	}

	return false
}
