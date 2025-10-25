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
)

const (
	GITHUB_REDIRECT_URI = "http://localhost:3000/api/auth/callback/github"
)

// Авторизация
func (h *Handler) SignUp(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Mail     string `json:"mail" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

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

	user, err := h.service.Authorization.CreateUser(user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": user.ID,
	})
}

type signInInput struct {
	Mail     string
	Password string
}

// Аунтефикация
func (h *Handler) SignIn(c *gin.Context) {
	var input signInInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	token, err := h.service.SignInWithPassword(input.Mail, input.Password)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// Методы обработки OAuth 2.0
func (h *Handler) RedirectToGithub(c *gin.Context) {
	redirect_url := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email",
		os.Getenv("GITHUB_CLIENT_ID"),
		GITHUB_REDIRECT_URI,
	)
	c.Redirect(http.StatusFound, redirect_url)
}

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
		"user": gin.H{
			"id":    user.ID,
			"email": user.Mail,
			"name":  user.Name,
		},
	})
}

// Обмен кода на access token
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
