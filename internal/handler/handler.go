package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/service"
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

	user := api.Group("/user")
	user.Use(h.userIdentity) // все эндпоинты требуют аутентификации
	{
		user.GET("/profile", h.GetUserProfile)
		user.PUT("/profile", h.UpdateUserProfile)
		user.PUT("/password", h.ChangePassword)
		user.DELETE("/account", h.DeleteAccount)
		user.POST("/github/link", h.LinkGitHubAccount)
		user.POST("/github/unlink", h.UnlinkGitHubAccount)
	}

	collectinons := api.Group("/collections")
	collectinons.Use(h.userIdentity)
	{
		collectinons.GET("", h.GetCollectionList) // User Collections, which him created
		collectinons.POST("", h.CreateCollection)
		// Get all collection items
		collectinons.GET("/:id/items", h.GetCollectionItems)
		// Add item to collection
		collectinons.POST("/:id/items", h.AddItemToCollection)
	}

	collectinon_items := api.Group("/items")
	collectinon_items.Use((h.userIdentity))
	{
		collectinon_items.GET("", h.GetItemsByType)
		collectinon_items.POST("", h.CreateCollectionItem)
	}

	return router

}
