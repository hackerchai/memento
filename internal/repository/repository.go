package repository

import (
	"go.uber.org/fx"
)

// Module exports repository related dependencies.
var Module = fx.Module("repository",
	fx.Provide(
		fx.Annotate(
			NewUserRepository,
			fx.ParamTags(``, ``),
		),
		// Add other repository providers here
	),
)
