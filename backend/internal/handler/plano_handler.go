package handler

import (
	"context"
	"database/sql"

	"github.com/gofiber/fiber/v2"

	"github.com/apsferreira/storemake/backend/internal/repository"
)

// PlanoHandler gerencia as rotas de planos
type PlanoHandler struct {
	planoRepo *repository.PlanoRepository
}

// NewPlanoHandler cria uma nova instância do handler
func NewPlanoHandler(db *sql.DB) *PlanoHandler {
	return &PlanoHandler{
		planoRepo: repository.NewPlanoRepository(db),
	}
}

// ListPricing retorna todos os planos ativos para exibição na LP
// GET /api/v1/public/plans
func (h *PlanoHandler) ListPricing(c *fiber.Ctx) error {
	ctx := context.Background()

	planos, err := h.planoRepo.ListActive(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "falha ao buscar planos"})
	}

	// Formatar resposta para LP
	type PriceDisplay struct {
		ID             string      `json:"id"`
		Slug           string      `json:"slug"`
		Name           string      `json:"name"`
		Price          string      `json:"price"`
		PriceCents     int         `json:"price_cents"`
		MaxProducts    int         `json:"max_products"`
		CustomDomain   bool        `json:"custom_domain"`
		SupportLevel   string      `json:"support_level"`
		Features       interface{} `json:"features"`
	}

	var response []PriceDisplay
	for _, p := range planos {
		var features interface{}
		if p.Features != nil && len(p.Features) > 0 {
			_ = p.Features.UnmarshalJSON(p.Features)
			features = p.Features
		}

		response = append(response, PriceDisplay{
			ID:           p.ID.String(),
			Slug:         p.Slug,
			Name:         p.Name,
			Price:        p.PriceDisplay(),
			PriceCents:   p.PriceCents,
			MaxProducts:  p.MaxProducts,
			CustomDomain: p.CustomDomain,
			SupportLevel: p.SupportLevel,
			Features:     features,
		})
	}

	return c.JSON(fiber.Map{
		"data":  response,
		"count": len(response),
	})
}

// GetBySlug retorna um plano específico
// GET /api/v1/public/plans/:slug
func (h *PlanoHandler) GetBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(400).JSON(fiber.Map{"error": "slug é obrigatório"})
	}

	ctx := context.Background()

	plano, err := h.planoRepo.GetBySlug(ctx, slug)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "falha ao buscar plano"})
	}

	if plano == nil {
		return c.Status(404).JSON(fiber.Map{"error": "plano não encontrado"})
	}

	return c.JSON(fiber.Map{"data": plano})
}
