package middleware

// SPEC-006-B: Middleware que verifica se um módulo está habilitado para o tenant.

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/repository"
)

// ModuleGuard retorna um middleware que bloqueia o acesso se o módulo estiver desabilitado.
// Extrai o store_id do JWT e consulta tenant_modules.
func ModuleGuard(module model.ModuleName) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "não autenticado",
			})
		}

		storeID, _ := claims["store_id"].(string)
		if storeID == "" {
			return c.Next() // sem store_id, deixa passar (não é tenant)
		}

		enabled, err := repository.IsModuleEnabled(storeID, module)
		if err != nil {
			// Em caso de erro de DB, permite acesso (fail-open para não degradar serviço)
			return c.Next()
		}

		if !enabled {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":  "módulo desabilitado",
				"module": string(module),
			})
		}

		return c.Next()
	}
}
