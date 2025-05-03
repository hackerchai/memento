package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// RootOnly is a middleware that ensures only root users can access the route.
func RootOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals(ContextUserRoleKey).(string)
		if role != "root" {
			return fiber.NewError(fiber.StatusForbidden, "Root access required")
		}
		return c.Next()
	}
}
