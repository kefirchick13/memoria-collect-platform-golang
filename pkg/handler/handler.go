package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kefirchick13/memoria-collect-platform-golang/pkg/service"
	"go.uber.org/zap"
)

type Handler struct {
	service *service.Service
	logger  *zap.SugaredLogger
}

func NewHandler(service *service.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	api := router.Group("/api")

	auth := api.Group("/auth")
	{
		auth.POST("/sign-up", h.SignUp)
		auth.POST("/sign-in", h.SignIn)
		auth.GET("/github", h.RedirectToGithub)
		auth.GET("/callback/github", h.GithubCallback)
	}

	return router

}
