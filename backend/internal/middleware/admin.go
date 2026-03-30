package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// AdminOnly verifica se o usuário tem role admin nas claims JWT.
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "acesso negado: claims não encontrados",
			})
		}

		role, _ := claims["role"].(string)
		if role != "admin" && role != "owner" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "acesso negado: requer permissão admin",
			})
		}

		return c.Next()
	}
}
