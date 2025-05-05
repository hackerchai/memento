package service

import (
	"time"

	"go.uber.org/fx"

	"github.com/hackerchai/memento/internal/auth"
	"github.com/hackerchai/memento/internal/config"
	"github.com/hackerchai/memento/internal/repository"
	"github.com/hackerchai/memento/pkg/xlog"
)

// Module exports dependency constructors for service implementations.
var Module = fx.Options(
	fx.Provide(
		// Provide NewUserService with correct dependencies
		func(userRepo *repository.UserRepository, appConfigRepo *repository.AppConfigRepository, jwtMaker *auth.JWTMaker, cfg *config.Config, logger *xlog.Logger) *UserService {
			// FIXME: Replace with actual config value for JWT duration
			// Example: tokenDuration := time.Duration(cfg.JWT.YourExpiryField) * time.Minute
			tokenDuration := 60 * time.Minute // Defaulting to 60 minutes
			return NewUserService(userRepo, appConfigRepo, jwtMaker, tokenDuration, logger)
		},
		NewArticleService, // This now implicitly requires AppConfigRepository
		NewAppConfigService,
	),
)
