package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hackerchai/memento/internal/auth"
	"github.com/hackerchai/memento/internal/config"
	"github.com/hackerchai/memento/internal/errmsg"
	"github.com/hackerchai/memento/internal/response"
	"github.com/hackerchai/memento/pkg/xlog"
)

const (
	// ContextUserIDKey is the key for storing user ID (uuid.UUID) in Fiber context.
	ContextUserIDKey = "ctx_user_id"
	// ContextUserRoleKey is the key for storing user role (string) in Fiber context.
	ContextUserRoleKey = "ctx_user_role"
)

var (
	// ErrMissingUserID is returned when user ID is not found in context.
	ErrMissingUserID = errors.New("user ID not found in context")
	// ErrMissingUserRole is returned when user role is not found in context.
	ErrMissingUserRole = errors.New("user role not found in context")
	// ErrInvalidUserIDType is returned when user ID in context is not uuid.UUID.
	ErrInvalidUserIDType = errors.New("invalid user ID type in context, expected uuid.UUID")
	// ErrInvalidUserRoleType is returned when user role in context is not string.
	ErrInvalidUserRoleType = errors.New("invalid user role type in context, expected string")
)

// AuthMiddleware creates a Fiber middleware for JWT authentication.
// It verifies the token and stores the UserID (uuid.UUID) and Role (string) in the context.
func AuthMiddleware(cfg *config.Config, logger *xlog.Logger) fiber.Handler {
	maker, err := auth.NewJWTMaker(cfg.JWT.Secret)
	if err != nil {
		logger.FatalX().Err(err).Msg("Failed to create JWTMaker due to invalid secret key config")
		return func(c *fiber.Ctx) error {
			return response.HandleError(c, logger, errmsg.ErrServer.WithDetails("JWT configuration error"))
		}
	}

	return func(c *fiber.Ctx) error {
		authHeader := c.Get(fiber.HeaderAuthorization)
		if authHeader == "" {
			logger.DebugX().Msg("Authorization header missing")
			return response.HandleError(c, logger, errmsg.ErrUnauthorized)
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || !strings.EqualFold(headerParts[0], "Bearer") {
			logger.WarnX().Str("header", authHeader).Msg("Invalid Authorization header format")
			return response.HandleError(c, logger, errmsg.ErrUnauthorized)
		}

		tokenString := headerParts[1]

		// Verify the token using the JWTMaker, expecting *auth.Payload
		payload, err := maker.VerifyToken(tokenString)
		if err != nil {
			logger.WarnX().Err(err).Str("token_prefix", tokenString[:min(len(tokenString), 10)]).Msg("Token verification failed")
			errStr := err.Error()
			if strings.Contains(errStr, "token has expired") {
				return response.HandleError(c, logger, errmsg.ErrUnauthorized.WithDetails("Token has expired"))
			}
			return response.HandleError(c, logger, errmsg.ErrUnauthorized.WithDetails("Token is invalid or malformed"))
		}

		// Extract UserID and Role from payload
		userID := payload.UserID
		role := payload.Role

		// Optional: Validate if UserID is nil UUID (should not happen if token is valid)
		if userID == uuid.Nil {
			logger.ErrorX().Msg("Valid token payload contained nil UserID")
			return response.HandleError(c, logger, errmsg.ErrUnauthorized.WithDetails("Invalid user identifier in token"))
		}

		// Store userID and role in context
		c.Locals(ContextUserIDKey, userID)
		c.Locals(ContextUserRoleKey, role)

		return c.Next()
	}
}

// GetUserIDFromContext retrieves the user ID (uuid.UUID) from the Fiber context.
// Returns uuid.Nil and an error if the user ID is not found or not the correct type.
func GetUserIDFromContext(c *fiber.Ctx, logger *xlog.Logger) (uuid.UUID, error) {
	userIDVal := c.Locals(ContextUserIDKey)
	if userIDVal == nil {
		return uuid.Nil, ErrMissingUserID
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		logger.ErrorX().Str("key", ContextUserIDKey).Type("found_type", userIDVal).Msg("Invalid type found for user ID in context")
		return uuid.Nil, ErrInvalidUserIDType
	}

	return userID, nil
}

// GetUserRoleFromContext retrieves the user role (string) from the Fiber context.
// Returns an empty string and an error if the role is not found or not the correct type.
func GetUserRoleFromContext(c *fiber.Ctx, logger *xlog.Logger) (string, error) {
	userRoleVal := c.Locals(ContextUserRoleKey)
	if userRoleVal == nil {
		return "", ErrMissingUserRole
	}

	userRole, ok := userRoleVal.(string)
	if !ok {
		logger.ErrorX().Str("key", ContextUserRoleKey).Type("found_type", userRoleVal).Msg("Invalid type found for user role in context")
		return "", ErrInvalidUserRoleType
	}

	return userRole, nil
}

// Helper for safe slicing
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
