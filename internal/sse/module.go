package sse

import (
	"go.uber.org/fx"
)

// Module provides the SSE broker to the fx container.
var Module = fx.Options(
	fx.Provide(
		// Provide the SSE Broker constructor.
		// It depends on *xlog.Logger, which should be provided elsewhere.
		NewBroker,
	),
)
