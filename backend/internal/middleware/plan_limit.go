package middleware

import (
	"context"
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/apsferreira/storemake/backend/internal/repository"
)

// PlanLimitValidator valida se a loja respeita o limite de produtos do seu plano
type PlanLimitValidator struct {
	planoRepo *repository.PlanoRepository
}

// NewPlanLimitValidator cria um novo validador de limites de plano
func NewPlanLimitValidator(db *sql.DB) *PlanLimitValidator {
	return &PlanLimitValidator{
		planoRepo: repository.NewPlanoRepository(db),
	}
}

// ValidateProductLimit é middleware que verifica se a loja pode criar mais produtos
func (pv *PlanLimitValidator) ValidateProductLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		lojaID, ok := c.Locals("lojaID").(string)
		if !ok {
			return c.Status(400).JSON(fiber.Map{"error": "loja_id ausente no contexto"})
		}

		lojaUUID, err := uuid.Parse(lojaID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "loja_id inválido"})
		}

		ctx := context.Background()

		// Verificar se pode criar mais produtos
		canCreate, err := pv.planoRepo.CanCreateProduct(ctx, lojaUUID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "falha ao verificar limite de produtos"})
		}

		if !canCreate {
			plano, err := pv.planoRepo.GetLojaPlano(ctx, lojaUUID)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "falha ao buscar plano"})
			}

			count, _ := pv.planoRepo.GetProductCount(ctx, lojaUUID)

			var planName string
			if plano != nil {
				planName = plano.Name
			}

			return c.Status(403).JSON(fiber.Map{
				"error":             "limite de produtos atingido",
				"plan":              planName,
				"current_count":     count,
				"max_products":      plano.MaxProducts,
				"upgrade_required":  true,
			})
		}

		return c.Next()
	}
}

// ValidateCustomDomain é middleware que verifica se a loja pode usar domínio customizado
func (pv *PlanLimitValidator) ValidateCustomDomain() fiber.Handler {
	return func(c *fiber.Ctx) error {
		lojaID, ok := c.Locals("lojaID").(string)
		if !ok {
			return c.Status(400).JSON(fiber.Map{"error": "loja_id ausente no contexto"})
		}

		lojaUUID, err := uuid.Parse(lojaID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "loja_id inválido"})
		}

		ctx := context.Background()

		plano, err := pv.planoRepo.GetLojaPlano(ctx, lojaUUID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "falha ao buscar plano"})
		}

		if plano == nil {
			return c.Status(404).JSON(fiber.Map{"error": "plano não encontrado"})
		}

		if !plano.CustomDomain {
			return c.Status(403).JSON(fiber.Map{
				"error":            "domínio customizado não permitido neste plano",
				"plan":             plano.Name,
				"upgrade_required": true,
			})
		}

		return c.Next()
	}
}
