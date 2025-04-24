package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"scripts-management/internal/models"
	"scripts-management/pkg/utils"
)

func AuthMiddleware(jwtManager *utils.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		c.Locals("user", claims)
		return c.Next()
	}
}

func RoleAuth(allowedRoles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*utils.JWTClaims)
		currentRole := models.UserRole(user.Role)

		if currentRole == models.RoleRoot {
			return c.Next()
		}

		for _, role := range allowedRoles {
			if currentRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}