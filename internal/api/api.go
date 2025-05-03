package api

import (
	"go.uber.org/fx"
)

// Module exports API handler related dependencies.
var Module = fx.Module("api",
	fx.Provide(
		fx.Annotate(
			NewAuthHandler,
			fx.ParamTags(``, ``),
		),
		// Add other handler providers here
		fx.Annotate(
			NewUserHandler,
			fx.ParamTags(``, ``),
		),
	),
)
