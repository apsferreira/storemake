package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/repository"
)

// ListOrders lista pedidos da loja (admin).
func ListOrders(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	page, perPage := paginationParams(c)
	filter := model.OrderListFilter{
		Page:    page,
		PerPage: perPage,
	}

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	orders, total, err := repository.ListOrders(storeID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao listar pedidos"})
	}

	if orders == nil {
		orders = []model.Order{}
	}

	totalPages := total / filter.PerPage
	if total%filter.PerPage != 0 {
		totalPages++
	}

	return c.JSON(model.PaginatedResponse{
		Data:       orders,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	})
}

// GetOrder retorna detalhes de um pedido (admin).
func GetOrder(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	order, err := repository.GetOrder(storeID, c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar pedido"})
	}
	if order == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "pedido não encontrado"})
	}

	return c.JSON(order)
}

// UpdateOrderStatus atualiza o status de um pedido (admin).
func UpdateOrderStatus(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req model.UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}

	validStatuses := map[string]bool{
		"pendente": true, "pago": true, "preparando": true,
		"enviado": true, "entregue": true, "cancelado": true,
	}
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if !validStatuses[status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "status inválido (pendente, pago, preparando, enviado, entregue, cancelado)",
		})
	}

	order, err := repository.UpdateOrderStatus(storeID, c.Params("id"), status, req.TrackingCode)
	if err != nil {
		if strings.Contains(err.Error(), "não encontrado") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		if strings.Contains(err.Error(), "transição") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao atualizar status"})
	}

	// BKL-144: notificação WhatsApp (non-fatal — não bloqueia resposta ao admin)
	NotifyOrderStatusViaWA(order.CustomerPhone, order.ID, status)

	return c.JSON(order)
}
