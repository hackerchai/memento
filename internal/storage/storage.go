package storage

import "go.uber.org/fx"

// Module exports dependency constructors for storage implementations.
var Module = fx.Options(
	// Provide LocalImageStorage as the implementation for ImageStorage interface.
	fx.Provide(
		fx.Annotate(
			NewLocalImageStorage,
			fx.As(new(ImageStorage)), // Provide as the ImageStorage interface
		),
	),
)
