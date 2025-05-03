package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hackerchai/memento/internal/config"
	"github.com/hackerchai/memento/pkg/xlog"
)

// XlogMiddleware creates a Fiber middleware for request logging and context injection.
// Renamed to XlogMiddleware for export.
func XlogMiddleware(logger *xlog.Logger, config *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Generate or get ReqID
		reqID := c.Get(string(xlog.DefaultReqIDKey))
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Set(string(xlog.DefaultReqIDKey), reqID)
		c.Request().Header.Set(string(xlog.DefaultReqIDKey), reqID)

		// Add logger and ReqID to context
		ctx := logger.WithContext(c.UserContext())
		ctx = logger.SetReqIDToContext(ctx, reqID)

		// Let's simplify, assuming WithContext is enough if the logger already has reqID
		c.SetUserContext(ctx) // Update Fiber's context

		// Log start of request
		logger.InfoX(ctx).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Str("userAgent", c.Get(fiber.HeaderUserAgent)).
			Msg("Request started")

		// Proceed with the request
		err := c.Next()

		if config.App.DebugMode {
			// Log end of request
			latency := time.Since(start)
			status := c.Response().StatusCode()

			logEvent := logger.InfoX(ctx) // Default to Info level for success
			if err != nil {
				// Error was handled by the global error handler, but we log it here too
				// The global handler already logged the error details
				logEvent = logger.ErrorX(ctx).Err(err) // Corrected call
				// We might already have the status code from the error handler response
				if e, ok := err.(*fiber.Error); ok {
					status = e.Code
				} else {
					// If it's not a fiber error, status might still be 2xx range
					// but we logged an error. Default to 500 if status is still 2xx.
					if status < 400 {
						status = fiber.StatusInternalServerError
					}
				}
			} else if status >= 400 {
				// Handle cases where Next() doesn't return an error but status is >= 400
				logEvent = logger.WarnX(ctx) // Use Warn for client/server errors without Go error
			}

			logEvent.
				Int("status", status).
				Str("latency", latency.String()).            // Human-readable latency
				Int64("latency_ms", latency.Milliseconds()). // Latency in ms
				Msg("Request finished")

		}

		return err // Return the error reported by c.Next()
	}
}
