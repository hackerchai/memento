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
	CategoryHandler  *api.CategoryHandler
	TagHandler       *api.TagHandler
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

	// Root Article routes (nested under root user routes)
	rootArticleGroup := rootGroup.Group("/articles")                     // --> /api/v1/users/root/articles
	rootArticleGroup.Get("/user", p.ArticleHandler.ListUserArticlesRoot) // GET /user?user_id=...
	rootArticleGroup.Get("/search", p.ArticleHandler.SearchArticlesRoot) // Search articles (root)
	rootArticleGroup.Get("/:id", p.ArticleHandler.GetArticleRoot)
	rootArticleGroup.Delete("/:id", p.ArticleHandler.DeleteArticleRoot)
	rootArticleGroup.Post("/:id/rescrape", p.ArticleHandler.ReScrapeArticleRoot)

	// --- Regular User Routes (Protected) ---

	// Article routes
	articleGroup := protected.Group("/articles") // --> /api/v1/articles
	articleGroup.Post("", p.ArticleHandler.CreateArticle)
	articleGroup.Get("", p.ArticleHandler.ListArticles)                         // List user's articles
	articleGroup.Get("/search", p.ArticleHandler.SearchArticles)                // Search user's articles
	articleGroup.Get("/:id", p.ArticleHandler.GetArticle)                       // Get specific article
	articleGroup.Delete("/:id", p.ArticleHandler.DeleteArticle)                 // Delete article
	articleGroup.Patch("/:id", p.ArticleHandler.UpdateArticleStatus)            // Update read/starred
	articleGroup.Post("/:id/rescrape", p.ArticleHandler.ReScrapeArticle)        // Re-scrape
	articleGroup.Post("/:id/tags", p.ArticleHandler.AddTagsToArticle)           // Add tags
	articleGroup.Delete("/:id/tags", p.ArticleHandler.RemoveTagsFromArticle)    // Remove tags
	articleGroup.Patch("/:id/category", p.ArticleHandler.UpdateArticleCategory) // Change category

	// SSE route
	sseGroup := protected.Group("/sse")
	sseGroup.Get("", p.SSEHandler.ConnectSSE)

	// App Config routes (for authenticated user)
	appConfigGroup := protected.Group("/config")
	appConfigGroup.Get("/", p.AppConfigHandler.GetOwnConfig)
	appConfigGroup.Put("/", p.AppConfigHandler.UpdateOwnConfig)

	// Category routes (for authenticated user)
	categoryGroup := protected.Group("/categories")
	categoryGroup.Post("", p.CategoryHandler.CreateCategory) // Create category
	categoryGroup.Get("", p.CategoryHandler.ListCategories)
	categoryGroup.Get("/search", p.CategoryHandler.SearchCategories) // Search user's categories
	categoryGroup.Get("/:identifier/articles", p.CategoryHandler.GetArticlesByCategory)
	categoryGroup.Delete("/:id", p.CategoryHandler.DeleteCategory)

	// Tag routes (for authenticated user)
	tagGroup := protected.Group("/tags")
	tagGroup.Post("", p.TagHandler.CreateTag) // Create tag
	tagGroup.Get("", p.TagHandler.ListTags)
	tagGroup.Get("/search", p.TagHandler.SearchTags) // Search user's tags
	tagGroup.Get("/:identifier/articles", p.TagHandler.GetArticlesByTag)
	tagGroup.Delete("/:id", p.TagHandler.DeleteTag)

	// --- Root Routes --- //
	// (RootOnly middleware is applied to rootGroup)

	// Root Category routes
	rootCategoriesGroup := rootGroup.Group("/categories")              // --> /api/v1/users/root/categories
	rootCategoriesGroup.Post("", p.CategoryHandler.CreateCategoryRoot) // Create category for target user
	rootCategoriesGroup.Get("", p.CategoryHandler.ListCategoriesRoot)  // Requires target_user_id query param
	rootCategoriesGroup.Get("/:identifier/articles", p.CategoryHandler.GetArticlesByCategoryRoot)
	rootCategoriesGroup.Delete("/:id", p.CategoryHandler.DeleteCategoryRoot)

	// Root Tag routes
	rootTagsGroup := rootGroup.Group("/tags")          // --> /api/v1/users/root/tags
	rootTagsGroup.Post("", p.TagHandler.CreateTagRoot) // Create tag for target user
	rootTagsGroup.Get("", p.TagHandler.ListTagsRoot)   // Requires target_user_id query param
	rootTagsGroup.Get("/:identifier/articles", p.TagHandler.GetArticlesByTagRoot)
	rootTagsGroup.Delete("/:id", p.TagHandler.DeleteTagRoot)
}
