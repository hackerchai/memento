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
		fx.Annotate(
			NewArticleHandler,
			fx.ParamTags(``, ``),
		),
		fx.Annotate(
			NewSSEHandler,
			fx.ParamTags(``, ``),
		),
		fx.Annotate(
			NewAppConfigHandler,
			fx.ParamTags(``, ``),
		),
		fx.Annotate(
			NewCategoryHandler,
			fx.ParamTags(``, ``),
		),
		fx.Annotate(
			NewTagHandler,
			fx.ParamTags(``, ``),
		),
	),
)
