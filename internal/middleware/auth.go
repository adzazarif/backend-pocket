package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"pocket-app/internal/pkg/apperror"
	"pocket-app/internal/pkg/jwt"
)

func Auth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return apperror.Unauthorized("Authorization header missing")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return apperror.Unauthorized("Invalid authorization header format")
		}

		tokenString := parts[1]
		claims, err := jwt.ParseToken(tokenString, secret)
		if err != nil {
			return err
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("name", claims.Name)

		return c.Next()
	}
}
