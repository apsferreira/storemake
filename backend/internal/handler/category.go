package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/repository"
)

func CreateCategory(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req model.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(req.Slug)

	if req.Name == "" || req.Slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name e slug são obrigatórios"})
	}
	if len(req.Name) > 255 || len(req.Slug) > 255 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name e slug devem ter até 255 caracteres"})
	}

	cat, err := repository.CreateCategory(storeID, req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "slug já existe nesta loja"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao criar categoria"})
	}

	return c.Status(fiber.StatusCreated).JSON(cat)
}

func ListCategories(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	categories, err := repository.ListCategories(storeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao listar categorias"})
	}

	if categories == nil {
		categories = []model.Category{}
	}

	return c.JSON(fiber.Map{"data": categories})
}

func GetCategory(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	cat, err := repository.GetCategory(storeID, c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar categoria"})
	}
	if cat == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "categoria não encontrada"})
	}

	return c.JSON(cat)
}

func UpdateCategory(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req model.UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" || len(trimmed) > 255 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name inválido"})
		}
		req.Name = &trimmed
	}
	if req.Slug != nil {
		trimmed := strings.TrimSpace(*req.Slug)
		if trimmed == "" || len(trimmed) > 255 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug inválido"})
		}
		req.Slug = &trimmed
	}

	cat, err := repository.UpdateCategory(storeID, c.Params("id"), req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "slug já existe nesta loja"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao atualizar categoria"})
	}
	if cat == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "categoria não encontrada"})
	}

	return c.JSON(cat)
}

func DeleteCategory(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	if err := repository.DeleteCategory(storeID, c.Params("id")); err != nil {
		if strings.Contains(err.Error(), "não encontrada") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "categoria não encontrada"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao deletar categoria"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
