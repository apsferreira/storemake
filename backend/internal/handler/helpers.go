package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// extractStoreID extrai o store_id das claims JWT do contexto.
// Espera que o JWT contenha claim "store_id".
func extractStoreID(c *fiber.Ctx) (string, error) {
	claims, ok := c.Locals("claims").(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("claims não encontrados no contexto")
	}

	storeID, ok := claims["store_id"].(string)
	if !ok || storeID == "" {
		return "", fmt.Errorf("store_id não encontrado no token")
	}

	return storeID, nil
}

func paginationParams(c *fiber.Ctx) (page, perPage int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	perPage, _ = strconv.Atoi(c.Query("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return
}
