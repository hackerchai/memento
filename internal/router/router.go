package router

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/hackerchai/memento/internal/api"
	"github.com/hackerchai/memento/internal/config"
	"github.com/hackerchai/memento/internal/middleware"
	"github.com/hackerchai/memento/pkg/xlog"
)

// Module exports router related dependencies.
var Module = fx.Module("router",
	fx.Provide( // Provide the RegisterRoutes function itself if needed elsewhere, or just Invoke it
	// No providers needed for now, just an Invoke target
	),
	fx.Invoke(RegisterRoutes), // Invoke the function to register routes during app start
)

// RegisterRoutesParams holds the dependencies for RegisterRoutes.
type RegisterRoutesParams struct {
	fx.In

	App              *fiber.App
	Config           *config.Config
	Logger           *xlog.Logger
	AuthHandler      *api.AuthHandler
	UserHandler      *api.UserHandler
	ArticleHandler   *api.ArticleHandler
	SSEHandler       *api.SSEHandler
	AppConfigHandler *api.AppConfigHandler
}

// RegisterRoutes registers all application routes.
func RegisterRoutes(p RegisterRoutesParams) {
	// --- Static File Serving ---
	// Serve cached images
	// TODO: Make the public prefix ("/assets/images") configurable?
	publicImagePrefix := "/assets/images"
	localImageBasePath := p.Config.Storage.Local.BasePath // Get base path from config
	if localImageBasePath == "" {
		localImageBasePath = "assets/images" // Fallback to default if not set
	}
	p.App.Static(publicImagePrefix, localImageBasePath, fiber.Static{
		Compress:  true,  // Enable compression
		ByteRange: true,  // Enable byte range requests
		Browse:    false, // Disable directory browsing
		MaxAge:    3600,  // Cache for 1 hour (adjust as needed)
	})
	p.Logger.InfoX(context.Background()).Str("prefix", publicImagePrefix).Str("path", localImageBasePath).Msg("Serving static images")

	// --- API Routes ---
	// Group routes
	apiGroup := p.App.Group("/api/v1")

	// Auth routes
	authGroup := apiGroup.Group("/auth")
	authGroup.Post("/register", p.AuthHandler.Register)
	authGroup.Post("/login", p.AuthHandler.Login)

	// Protected routes (require auth middleware)
	protected := apiGroup.Group("", middleware.AuthMiddleware(p.Config, p.Logger))

	// User routes
	userGroup := protected.Group("/users")
	userGroup.Get("/profile", p.UserHandler.GetProfile)
	userGroup.Put("/password", p.UserHandler.UpdatePassword)
	userGroup.Put("/email", p.UserHandler.UpdateEmail)
	userGroup.Put("/name", p.UserHandler.UpdateName)

	// Root only user routes
	rootGroup := userGroup.Group("/root")
	rootGroup.Use(middleware.RootOnly())
	rootGroup.Get("/", p.UserHandler.GetUserList)
	rootGroup.Post("/", p.UserHandler.CreateUser)
	rootGroup.Delete("/:id", p.UserHandler.DeleteUser)

	// Root AppConfig routes (nested under root user routes for clarity)
	rootConfigGroup := rootGroup.Group("/config")
	rootConfigGroup.Get("/", p.AppConfigHandler.ListConfigsRoot)
	rootConfigGroup.Get("/:id", p.AppConfigHandler.GetConfigByUserIDRoot)    // :id is target user ID
	rootConfigGroup.Put("/:id", p.AppConfigHandler.UpdateConfigByUserIDRoot) // :id is target user ID

	// Article routes
	articleGroup := protected.Group("/articles")
	articleGroup.Post("", p.ArticleHandler.CreateArticle)
	// TODO: Add other article routes (GET, DELETE etc.) later

	// SSE route
	sseGroup := protected.Group("/sse")
	sseGroup.Get("", p.SSEHandler.ConnectSSE)

	// App Config routes (for authenticated user)
	appConfigGroup := protected.Group("/config")
	appConfigGroup.Get("/", p.AppConfigHandler.GetOwnConfig)
	appConfigGroup.Put("/", p.AppConfigHandler.UpdateOwnConfig)
}
