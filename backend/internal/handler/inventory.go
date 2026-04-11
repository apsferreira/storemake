package handler

// BKL-900: Handlers de inventário multi-loja centralizado.
// CRUD de SKUs, alocações por loja, movimentações e alertas de reposição.

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/repository"
)

// ListInventoryMasters godoc
// GET /api/v1/inventory
// Lista todos os SKUs centralizados do tenant autenticado.
func ListInventoryMasters(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	masters, err := repository.ListInventoryMasters(storeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao listar estoque"})
	}
	return c.JSON(fiber.Map{"items": masters, "total": len(masters)})
}

// GetInventoryMaster godoc
// GET /api/v1/inventory/:id
// Retorna um SKU centralizado com suas alocações por loja.
func GetInventoryMaster(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	masterID := c.Params("id")
	master, err := repository.GetInventoryMaster(storeID, masterID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar SKU"})
	}
	if master == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "SKU não encontrado"})
	}

	allocations, _ := repository.ListStoreAllocations(masterID)
	movements, _ := repository.ListMovements(masterID, 20)

	return c.JSON(fiber.Map{
		"master":      master,
		"allocations": allocations,
		"movements":   movements,
	})
}

// CreateInventoryMaster godoc
// POST /api/v1/inventory
// Cria um novo SKU centralizado.
func CreateInventoryMaster(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	tenantUUID, err := uuid.Parse(storeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "store_id inválido"})
	}

	var req model.CreateInventoryMasterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if req.Nome == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "nome é obrigatório"})
	}
	if req.Unidade == "" {
		req.Unidade = "un"
	}
	if req.ReorderPoint <= 0 {
		req.ReorderPoint = 5
	}
	if req.ReorderQuantity <= 0 {
		req.ReorderQuantity = 10
	}
	req.TenantID = tenantUUID

	master, err := repository.CreateInventoryMaster(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao criar SKU"})
	}
	return c.Status(fiber.StatusCreated).JSON(master)
}

// AdjustInventoryQuantity godoc
// POST /api/v1/inventory/:id/adjust
// Ajusta a quantidade de um SKU (entrada, saída, ajuste manual).
func AdjustInventoryQuantity(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	masterID := c.Params("id")
	master, err := repository.GetInventoryMaster(storeID, masterID)
	if err != nil || master == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "SKU não encontrado"})
	}

	var req model.AdjustQuantityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if req.Delta == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "delta não pode ser zero"})
	}

	movType := model.MovementTypeAjuste
	if req.Delta > 0 {
		movType = model.MovementTypeEntrada
	} else {
		movType = model.MovementTypeSaidaPerda
	}

	if err := repository.UpdateInventoryMasterQuantity(masterID, req.Delta, movType, req.LojaID, req.Observacao); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao ajustar estoque"})
	}

	updated, _ := repository.GetInventoryMaster(storeID, masterID)
	return c.JSON(updated)
}

// UpsertStoreAllocation godoc
// PUT /api/v1/inventory/:id/allocations/:loja_id
// Cria ou atualiza a cota de estoque para uma loja.
func UpsertStoreAllocation(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	masterID := c.Params("id")
	lojaID := c.Params("loja_id")

	master, err := repository.GetInventoryMaster(storeID, masterID)
	if err != nil || master == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "SKU não encontrado"})
	}

	var req model.UpsertAllocationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if req.QuantityAllocated < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "quantidade não pode ser negativa"})
	}
	if req.ProfitSharePct < 0 || req.ProfitSharePct > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "profit_share_pct deve estar entre 0 e 100"})
	}

	allocation, err := repository.UpsertStoreAllocation(masterID, lojaID, req.QuantityAllocated, req.ProfitSharePct)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao atualizar alocação"})
	}
	return c.JSON(allocation)
}

// ListInventoryAlerts godoc
// GET /api/v1/inventory/alerts
// Lista alertas de estoque baixo não confirmados do tenant.
func ListInventoryAlerts(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	alerts, err := repository.ListPendingAlerts(storeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao listar alertas"})
	}
	return c.JSON(fiber.Map{"alerts": alerts, "count": len(alerts)})
}

// AcknowledgeInventoryAlert godoc
// POST /api/v1/inventory/alerts/:id/acknowledge
// Confirma que o alerta foi visto/tratado.
func AcknowledgeInventoryAlert(c *fiber.Ctx) error {
	alertID := c.Params("id")
	userID := c.Locals("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr, _ = userID.(string)
	}

	if err := repository.AcknowledgeAlert(alertID, userIDStr); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao confirmar alerta"})
	}
	return c.JSON(fiber.Map{"message": "alerta confirmado"})
}

// ListInventoryMovements godoc
// GET /api/v1/inventory/:id/movements?limit=50
// Lista as movimentações de um SKU para auditoria.
func ListInventoryMovements(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	masterID := c.Params("id")
	master, err := repository.GetInventoryMaster(storeID, masterID)
	if err != nil || master == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "SKU não encontrado"})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	movements, err := repository.ListMovements(masterID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao listar movimentações"})
	}
	return c.JSON(fiber.Map{"movements": movements, "count": len(movements)})
}

// CreateSupplierOrder godoc
// POST /api/v1/inventory/:id/orders
// Cria um pedido de reposição para um SKU.
func CreateSupplierOrder(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	masterID := c.Params("id")
	master, err := repository.GetInventoryMaster(storeID, masterID)
	if err != nil || master == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "SKU não encontrado"})
	}

	var req model.CreateSupplierOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if req.QuantityOrdered <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "quantity_ordered deve ser maior que zero"})
	}

	order, err := repository.CreateSupplierOrder(storeID, masterID, req.QuantityOrdered, req.CustoTotalCents, req.FornecedorNome, req.Observacao)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao criar pedido de reposição"})
	}
	return c.Status(fiber.StatusCreated).JSON(order)
}

// ListSupplierOrders godoc
// GET /api/v1/inventory/orders
// Lista pedidos de reposição do tenant.
func ListSupplierOrders(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	orders, err := repository.ListSupplierOrders(storeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao listar pedidos"})
	}
	return c.JSON(fiber.Map{"orders": orders, "count": len(orders)})
}

// --- Rotas /api/v1/inventory/items (BKL-900 spec canônica) ---

// ListInventoryItems godoc
// GET /api/v1/inventory/items
// Lista SKUs do tenant autenticado.
func ListInventoryItems(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	items, err := repository.ListInventoryMasters(storeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao listar estoque"})
	}
	return c.JSON(fiber.Map{"items": items, "total": len(items)})
}

// CreateInventoryItem godoc
// POST /api/v1/inventory/items
// Cria um novo SKU centralizado.
func CreateInventoryItem(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	tenantUUID, err := uuid.Parse(storeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "store_id inválido"})
	}

	var req model.CreateInventoryMasterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if strings.TrimSpace(req.Nome) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "nome é obrigatório"})
	}
	if req.Unidade == "" {
		req.Unidade = "un"
	}
	if req.ReorderPoint <= 0 {
		req.ReorderPoint = 5
	}
	if req.ReorderQuantity <= 0 {
		req.ReorderQuantity = 10
	}
	req.TenantID = tenantUUID

	item, err := repository.CreateInventoryMaster(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao criar SKU"})
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

// GetInventoryItem godoc
// GET /api/v1/inventory/items/:id
// Detalhe de um SKU com alocações e movimentações recentes.
func GetInventoryItem(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	itemID := c.Params("id")
	item, err := repository.GetInventoryMaster(storeID, itemID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar SKU"})
	}
	if item == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "SKU não encontrado"})
	}

	allocations, _ := repository.ListStoreAllocations(itemID)
	movements, _ := repository.ListMovements(itemID, 20)

	return c.JSON(fiber.Map{
		"item":        item,
		"allocations": allocations,
		"movements":   movements,
	})
}

// UpdateInventoryItem godoc
// PUT /api/v1/inventory/items/:id
// Atualiza os campos cadastrais de um SKU.
func UpdateInventoryItem(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	itemID := c.Params("id")

	var req model.UpdateInventoryMasterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if strings.TrimSpace(req.Nome) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "nome é obrigatório"})
	}
	if req.Unidade == "" {
		req.Unidade = "un"
	}

	updated, err := repository.UpdateInventoryMaster(itemID, storeID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao atualizar SKU"})
	}
	if updated == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "SKU não encontrado"})
	}
	return c.JSON(updated)
}

// DeleteInventoryItem godoc
// DELETE /api/v1/inventory/items/:id
// Remove (soft-delete) um SKU do tenant.
func DeleteInventoryItem(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	itemID := c.Params("id")
	if err := repository.DeleteInventoryMaster(itemID, storeID); err != nil {
		if strings.Contains(err.Error(), "não encontrado") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "SKU não encontrado"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao remover SKU"})
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// AllocateInventoryItem godoc
// POST /api/v1/inventory/items/:id/allocate
// Cria ou atualiza a alocação de um SKU para uma loja.
func AllocateInventoryItem(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	itemID := c.Params("id")
	item, err := repository.GetInventoryMaster(storeID, itemID)
	if err != nil || item == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "SKU não encontrado"})
	}

	var req model.UpsertAllocationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if req.LojaID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "loja_id é obrigatório"})
	}
	if req.QuantityAllocated < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "quantidade não pode ser negativa"})
	}
	if req.ProfitSharePct < 0 || req.ProfitSharePct > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "profit_share_pct deve estar entre 0 e 100"})
	}

	allocation, err := repository.UpsertStoreAllocation(itemID, req.LojaID, req.QuantityAllocated, req.ProfitSharePct)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao alocar estoque"})
	}
	return c.JSON(allocation)
}

// RegisterInventoryMovement godoc
// POST /api/v1/inventory/items/:id/movement
// Registra movimentação de estoque (entrada, saída, ajuste, devolução).
func RegisterInventoryMovement(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	itemID := c.Params("id")
	item, err := repository.GetInventoryMaster(storeID, itemID)
	if err != nil || item == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "SKU não encontrado"})
	}

	var req model.RegisterMovementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if req.MovementType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "movement_type é obrigatório"})
	}
	if req.Delta == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "delta não pode ser zero"})
	}

	// Para saídas: verificar se há estoque suficiente
	if req.Delta < 0 {
		available := item.QuantityAvailable()
		needed := -req.Delta
		if available < needed {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":     "estoque insuficiente",
				"available": available,
				"requested": needed,
			})
		}
	}

	if err := repository.RegisterMovement(itemID, storeID, req); err != nil {
		if strings.Contains(err.Error(), "insuficiente") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao registrar movimentação"})
	}

	updated, _ := repository.GetInventoryMaster(storeID, itemID)
	return c.JSON(fiber.Map{"item": updated, "message": "movimentação registrada"})
}

// ListAllInventoryMovements godoc
// GET /api/v1/inventory/movements
// Histórico geral de movimentações do tenant. Query param: limit (max 500).
func ListAllInventoryMovements(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	movements, err := repository.ListAllMovements(storeID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao listar movimentações"})
	}
	return c.JSON(fiber.Map{"movements": movements, "count": len(movements)})
}
