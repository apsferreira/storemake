package handler

// SPEC-006-B: Handlers para gerenciamento de feature flags de módulos.

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/repository"
)

// ListModules godoc
// GET /api/v1/modules
// Retorna todos os módulos do tenant autenticado com seu estado.
func ListModules(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	modules, err := repository.GetTenantModules(storeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar módulos"})
	}

	// Se não há módulos ainda, inicializa com padrão
	if len(modules) == 0 {
		if err := repository.EnsureTenantModules(storeID); err == nil {
			modules, _ = repository.GetTenantModules(storeID)
		}
	}

	// Converte config JSONB para map para resposta legível
	type moduleResponse struct {
		Module    string         `json:"module"`
		Enabled   bool           `json:"enabled"`
		Config    map[string]any `json:"config"`
		UpdatedAt string         `json:"updated_at"`
	}

	result := make([]moduleResponse, 0, len(modules))
	for _, m := range modules {
		var cfg map[string]any
		_ = json.Unmarshal(m.Config, &cfg)
		if cfg == nil {
			cfg = map[string]any{}
		}
		result = append(result, moduleResponse{
			Module:    m.Module,
			Enabled:   m.Enabled,
			Config:    cfg,
			UpdatedAt: m.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return c.JSON(fiber.Map{"modules": result})
}

// UpdateModule godoc
// PUT /api/v1/modules/:module
// Ativa ou desativa um módulo do tenant.
func UpdateModule(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	moduleName := c.Params("module")

	// Valida se o módulo é reconhecido
	validModule := false
	for _, m := range model.AllModules {
		if string(m) == moduleName {
			validModule = true
			break
		}
	}
	if !validModule {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "módulo desconhecido",
			"valid":   []string{"storefront", "inventory", "crm", "whatsapp"},
			"module":  moduleName,
		})
	}

	var req model.UpdateModuleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}

	if req.Config == nil {
		req.Config = map[string]any{}
	}

	if err := repository.UpsertTenantModule(storeID, model.ModuleName(moduleName), req.Enabled, req.Config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao atualizar módulo"})
	}

	return c.JSON(fiber.Map{
		"module":  moduleName,
		"enabled": req.Enabled,
		"message": "módulo atualizado com sucesso",
	})
}
