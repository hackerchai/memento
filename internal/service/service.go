package service

import (
	"go.uber.org/fx"
)

// Module exports service related dependencies.
var Module = fx.Module("service",
	fx.Provide(
		// Annotate NewUserService to inject dependencies including logger
		fx.Annotate(
			NewUserService,
			// Add tags if needed, e.g., for named dependencies
			// Params: UserRepository, JWTMaker, time.Duration, *xlog.Logger
			fx.ParamTags(``, ``, ``, ``),
		),
		// Add other service providers here
	),
)
