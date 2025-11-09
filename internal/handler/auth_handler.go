package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/models"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/responses"
)

const (
	GITHUB_REDIRECT_URI = "http://localhost:3000/api/auth/callback/github"
)

type signUpInput struct {
	Name     string `json:"name" binding:"required"`
	Mail     string `json:"mail" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SignUp создает нового пользователя
// @Summary Регистрация пользователя
// @Description Создает нового пользователя в системе
// @Tags auth
// @Accept json
// @Produce json
// @Param input body signUpInput true "Данные для регистрации"
// @Success 201 {object} models.UserResponse "Пользователь создан"
// @Failure 400 {object} responses.ErrorResponse "Ошибка валидации"
// @Router /auth/signup [post]
func (h *Handler) SignUp(c *gin.Context) {
	var input signUpInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &models.User{
		Name:      input.Name,
		Mail:      input.Mail,
		Password:  &input.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	newUser, err := h.service.AuthService.Register(user)

	if err != nil {
		responses.NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusCreated, newUser.ToResponse())
}

type signInInput struct {
	Mail     string `json:"mail" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Response structs for Swagger documentation
type signInResponse struct {
	Token string              `json:"token"`
	User  models.UserResponse `json:"user"`
}

// SignIn аутентифицирует пользователя
// @Summary Вход в систему
// @Description Аутентификация пользователя по email и паролю
// @Tags auth
// @Accept json
// @Produce json
// @Param input body signInInput true "Данные для входа"
// @Success 200 {object} signInResponse "Успешная аутентификация"
// @Failure 400 {object} responses.ErrorResponse "Неверные учетные данные"
// @Router /auth/signin [post]
func (h *Handler) SignIn(c *gin.Context) {
	var input signInInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, token, err := h.service.SignInWithPassword(input.Mail, input.Password)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, signInResponse{
		Token: token,
		User:  user.ToResponse(),
	})
}

// RedirectToGithub перенаправляет на GitHub OAuth
// @Summary Перенаправление на GitHub OAuth
// @Description Инициирует процесс OAuth аутентификации через GitHub
// @Tags auth
// @Success 302 "Перенаправление на GitHub"
// @Router /auth/github [get]
func (h *Handler) RedirectToGithub(c *gin.Context) {
	redirect_url := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email",
		os.Getenv("GITHUB_CLIENT_ID"),
		GITHUB_REDIRECT_URI,
	)
	c.Redirect(http.StatusFound, redirect_url)
}

// GithubCallback обрабатывает callback от GitHub OAuth
// @Summary OAuth callback от GitHub
// @Description Обрабатывает callback от GitHub после аутентификации
// @Tags auth
// @Produce json
// @Param code query string true "Код авторизации от GitHub"
// @Success 200 {object} signInResponse "Успешная аутентификация"
// @Failure 400 {object} responses.ErrorResponse "Отсутствует код авторизации"
// @Failure 500 {object} responses.ErrorResponse "Ошибка сервера"
// @Router /auth/callback/github [get]
func (h *Handler) GithubCallback(c *gin.Context) {
	code := c.Query("code")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "code not found",
		})
		return
	}

	// 1. Обмен кода на access token
	githubAccessToken, err := h.exchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to exchange code for token",
		})
		return
	}

	// 2. Получение данных пользователя из GitHub
	githubUser, err := h.getGitHubUser(githubAccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get user data from GitHub",
		})
		return
	}

	// 3. Поиск или создание пользователя в вашей БД
	user, err := h.service.FindOrCreateGitHubUser(githubUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to process user",
		})
		return
	}

	// 4. Генерация JWT токена для вашего приложения
	appToken, err := h.service.SignInWithOAuth(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate app token",
		})
		return
	}

	// 5. Возврат токена клиенту
	c.JSON(http.StatusOK, gin.H{
		"token": appToken,
		"user":  user,
	})
}

// Обмен кода из github на access token
func (h *Handler) exchangeCodeForToken(code string) (string, error) {
	tokenURL := "https://github.com/login/oauth/access_token"

	requestBody := map[string]string{
		"client_id":     os.Getenv("GITHUB_CLIENT_ID"),
		"client_secret": os.Getenv("GITHUB_CLIENT_SECRET"),
		"code":          code,
		"redirect_uri":  GITHUB_REDIRECT_URI,
	}

	jsonBody, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

// Получение данных пользователя из GitHub
func (h *Handler) getGitHubUser(accessToken string) (*models.GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var githubUser models.GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, err
	}

	marshaled, err := json.Marshal(githubUser)
	if err != nil {
		h.logger.Error("failed to marshal githubUser", err)
	} else {
		h.logger.Infof("githubUser: %s", marshaled)
	}

	return &githubUser, nil
}
