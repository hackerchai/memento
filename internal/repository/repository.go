package repository

import (
	"go.uber.org/fx"
)

// Module exports dependency constructors for repository implementations.
var Module = fx.Options(
	fx.Provide(NewUserRepository),
	fx.Provide(NewCategoryRepository),
	fx.Provide(NewTagRepository),
	fx.Provide(NewArticleRepository),
	fx.Provide(NewAppConfigRepository),
)
