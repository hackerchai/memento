package auth

import (
	"time"

	"go.uber.org/fx"

	"github.com/hackerchai/memento/internal/config"
)

// Module exports authentication related dependencies.
var Module = fx.Module("auth",
	fx.Provide(
		// Provide JWTMaker using config
		func(cfg *config.Config) (*JWTMaker, error) {
			return NewJWTMaker(cfg.JWT.Secret)
		},
		// Provide Access Token Duration using config
		func(cfg *config.Config) (time.Duration, error) {
			return cfg.JWT.AccessTokenDuration, nil
		},
		// Password hashing functions are typically used directly, no provider needed unless wrapped
	),
)
