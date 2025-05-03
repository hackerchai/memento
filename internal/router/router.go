package router

import (
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

	App         *fiber.App
	Config      *config.Config
	Logger      *xlog.Logger
	AuthHandler *api.AuthHandler
	UserHandler *api.UserHandler
}

// RegisterRoutes registers all application routes.
func RegisterRoutes(p RegisterRoutesParams) {
	// Group routes
	apiGroup := p.App.Group("/api/v1")

	// Auth routes
	authGroup := apiGroup.Group("/auth")
	authGroup.Post("/register", p.AuthHandler.Register)
	authGroup.Post("/login", p.AuthHandler.Login)

	// User routes (protected)
	userGroup := apiGroup.Group("/users")
	userGroup.Use(middleware.AuthMiddleware(p.Config, p.Logger))

	// Routes for all authenticated users
	userGroup.Get("/profile", p.UserHandler.GetProfile)
	userGroup.Put("/password", p.UserHandler.UpdatePassword)
	userGroup.Put("/email", p.UserHandler.UpdateEmail)
	userGroup.Put("/name", p.UserHandler.UpdateName)

	// Routes for root users only
	rootGroup := userGroup.Group("/root")
	rootGroup.Use(middleware.RootOnly())
	rootGroup.Get("/", p.UserHandler.GetUserList)
	rootGroup.Post("/", p.UserHandler.CreateUser)
	rootGroup.Delete("/:id", p.UserHandler.DeleteUser)
}
