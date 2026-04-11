package repository

// BKL-900: Repository de inventário multi-loja centralizado.
// Gerencia inventory_masters, store_allocations, inventory_movements,
// inventory_alerts e supplier_orders.

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/pkg/database"
)

// --- InventoryMaster ---

// ListInventoryMasters lista todos os SKUs de um tenant.
func ListInventoryMasters(tenantID string) ([]*model.InventoryMaster, error) {
	rows, err := database.DB.Query(`
		SELECT id, tenant_id, produto_id, sku_global, nome, descricao, unidade,
		       custo_unitario_cents, quantity_total, quantity_reserved,
		       reorder_point, reorder_quantity, is_active, created_at, updated_at
		FROM inventory_masters
		WHERE tenant_id = $1
		ORDER BY nome ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("inventory: list: %w", err)
	}
	defer rows.Close()

	var result []*model.InventoryMaster
	for rows.Next() {
		m, err := scanInventoryMaster(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

// GetInventoryMaster busca um SKU por ID, verificando pertencimento ao tenant.
func GetInventoryMaster(tenantID, masterID string) (*model.InventoryMaster, error) {
	row := database.DB.QueryRow(`
		SELECT id, tenant_id, produto_id, sku_global, nome, descricao, unidade,
		       custo_unitario_cents, quantity_total, quantity_reserved,
		       reorder_point, reorder_quantity, is_active, created_at, updated_at
		FROM inventory_masters
		WHERE id = $1 AND tenant_id = $2
	`, masterID, tenantID)
	return scanInventoryMasterRow(row)
}

// CreateInventoryMaster cria um novo SKU centralizado.
func CreateInventoryMaster(req model.CreateInventoryMasterRequest) (*model.InventoryMaster, error) {
	id := uuid.New()
	now := time.Now()

	_, err := database.DB.Exec(`
		INSERT INTO inventory_masters (
			id, tenant_id, produto_id, sku_global, nome, descricao, unidade,
			custo_unitario_cents, quantity_total, reorder_point, reorder_quantity,
			is_active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,true,$12,$13)
	`,
		id, req.TenantID, req.ProdutoID, req.SKUGlobal, req.Nome, req.Descricao,
		req.Unidade, req.CustoUnitarioCents, req.QuantityTotal,
		req.ReorderPoint, req.ReorderQuantity, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("inventory: create master: %w", err)
	}
	return GetInventoryMaster(req.TenantID.String(), id.String())
}

// UpdateInventoryMaster atualiza os campos cadastrais de um SKU (nome, descrição, reorder_point, etc.).
func UpdateInventoryMaster(masterID, tenantID string, req model.UpdateInventoryMasterRequest) (*model.InventoryMaster, error) {
	_, err := database.DB.Exec(`
		UPDATE inventory_masters
		SET nome = $1, descricao = $2, unidade = $3, custo_unitario_cents = $4,
		    reorder_point = $5, reorder_quantity = $6, sku_global = $7, updated_at = NOW()
		WHERE id = $8 AND tenant_id = $9
	`,
		req.Nome, req.Descricao, req.Unidade, req.CustoUnitarioCents,
		req.ReorderPoint, req.ReorderQuantity, req.SKUGlobal,
		masterID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("inventory: update master: %w", err)
	}
	return GetInventoryMaster(tenantID, masterID)
}

// DeleteInventoryMaster desativa (soft-delete) um SKU master do tenant.
func DeleteInventoryMaster(masterID, tenantID string) error {
	res, err := database.DB.Exec(`
		UPDATE inventory_masters
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`, masterID, tenantID)
	if err != nil {
		return fmt.Errorf("inventory: delete master: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("inventory: SKU não encontrado")
	}
	return nil
}

// DecrementStock decrementa estoque de um SKU numa loja específica dentro de uma transação.
// Retorna erro se quantity_available < qty.
func DecrementStock(masterID, lojaID string, qty int, observacao string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("inventory: begin tx: %w", err)
	}
	defer tx.Rollback()

	var quantityTotal int
	err = tx.QueryRow(`
		SELECT quantity_total FROM inventory_masters
		WHERE id = $1 FOR UPDATE
	`, masterID).Scan(&quantityTotal)
	if err != nil {
		return fmt.Errorf("inventory: lock master: %w", err)
	}

	if quantityTotal < qty {
		return fmt.Errorf("estoque insuficiente: disponível %d, solicitado %d", quantityTotal, qty)
	}

	before := quantityTotal
	after := quantityTotal - qty

	_, err = tx.Exec(`
		UPDATE inventory_masters
		SET quantity_total = quantity_total - $1, updated_at = NOW()
		WHERE id = $2
	`, qty, masterID)
	if err != nil {
		return fmt.Errorf("inventory: decrement quantity: %w", err)
	}

	var lojaUUID *uuid.UUID
	if lojaID != "" {
		u, err2 := uuid.Parse(lojaID)
		if err2 == nil {
			lojaUUID = &u
		}
	}

	_, err = tx.Exec(`
		INSERT INTO inventory_movements (
			id, master_id, loja_id, movement_type, quantity,
			quantity_before, quantity_after, observacao, created_at
		) VALUES ($1,$2,$3,'saida_venda',$4,$5,$6,$7,NOW())
	`, uuid.New(), masterID, lojaUUID, -qty, before, after, observacao)
	if err != nil {
		return fmt.Errorf("inventory: insert saida movement: %w", err)
	}

	return tx.Commit()
}

// ListAllMovements lista movimentações de todo o tenant com paginação.
func ListAllMovements(tenantID string, limit int) ([]*model.InventoryMovement, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := database.DB.Query(`
		SELECT mv.id, mv.master_id, mv.loja_id, mv.pedido_id, mv.movement_type,
		       mv.quantity, mv.quantity_before, mv.quantity_after,
		       mv.custo_unitario_cents, mv.observacao, mv.created_by, mv.created_at
		FROM inventory_movements mv
		JOIN inventory_masters m ON m.id = mv.master_id
		WHERE m.tenant_id = $1
		ORDER BY mv.created_at DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("inventory: list all movements: %w", err)
	}
	defer rows.Close()

	var result []*model.InventoryMovement
	for rows.Next() {
		mv, err := scanMovement(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, mv)
	}
	return result, rows.Err()
}

// RegisterMovement registra uma movimentação explícita (entrada/ajuste/devolução) e atualiza quantity_total.
func RegisterMovement(masterID, tenantID string, req model.RegisterMovementRequest) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("inventory: begin tx: %w", err)
	}
	defer tx.Rollback()

	var before int
	err = tx.QueryRow(`
		SELECT quantity_total FROM inventory_masters
		WHERE id = $1 AND tenant_id = $2 FOR UPDATE
	`, masterID, tenantID).Scan(&before)
	if err != nil {
		return fmt.Errorf("inventory: lock master for movement: %w", err)
	}

	after := before + req.Delta

	_, err = tx.Exec(`
		UPDATE inventory_masters
		SET quantity_total = quantity_total + $1, updated_at = NOW()
		WHERE id = $2
	`, req.Delta, masterID)
	if err != nil {
		return fmt.Errorf("inventory: apply movement delta: %w", err)
	}

	var lojaUUID *uuid.UUID
	if req.LojaID != "" {
		u, err2 := uuid.Parse(req.LojaID)
		if err2 == nil {
			lojaUUID = &u
		}
	}

	_, err = tx.Exec(`
		INSERT INTO inventory_movements (
			id, master_id, loja_id, movement_type, quantity,
			quantity_before, quantity_after, observacao, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
	`, uuid.New(), masterID, lojaUUID, req.MovementType, req.Delta, before, after, req.Observacao)
	if err != nil {
		return fmt.Errorf("inventory: insert movement: %w", err)
	}

	return tx.Commit()
}

// UpdateInventoryMasterQuantity atualiza quantity_total e registra a movimentação.
// Chamado automaticamente em vendas e recebimentos.
func UpdateInventoryMasterQuantity(masterID string, delta int, movType model.MovementType, lojaID, observacao string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("inventory: begin tx: %w", err)
	}
	defer tx.Rollback()

	var before, after int
	err = tx.QueryRow(`
		UPDATE inventory_masters
		SET quantity_total = quantity_total + $1, updated_at = NOW()
		WHERE id = $2
		RETURNING quantity_total - $1, quantity_total
	`, delta, masterID).Scan(&before, &after)
	if err != nil {
		return fmt.Errorf("inventory: update quantity: %w", err)
	}

	var lojaUUID *uuid.UUID
	if lojaID != "" {
		u, err := uuid.Parse(lojaID)
		if err == nil {
			lojaUUID = &u
		}
	}

	_, err = tx.Exec(`
		INSERT INTO inventory_movements (
			id, master_id, loja_id, movement_type, quantity,
			quantity_before, quantity_after, observacao, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
	`, uuid.New(), masterID, lojaUUID, movType, delta, before, after, observacao)
	if err != nil {
		return fmt.Errorf("inventory: insert movement: %w", err)
	}

	return tx.Commit()
}

// --- StoreAllocation ---

// ListStoreAllocations lista as alocações de um SKU master.
func ListStoreAllocations(masterID string) ([]*model.StoreAllocation, error) {
	rows, err := database.DB.Query(`
		SELECT id, master_id, loja_id, quantity_allocated, quantity_sold,
		       profit_share_pct, is_active, created_at, updated_at
		FROM store_allocations
		WHERE master_id = $1 AND is_active = true
		ORDER BY created_at ASC
	`, masterID)
	if err != nil {
		return nil, fmt.Errorf("inventory: list allocations: %w", err)
	}
	defer rows.Close()

	var result []*model.StoreAllocation
	for rows.Next() {
		a, err := scanAllocation(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

// UpsertStoreAllocation cria ou atualiza a alocação de uma loja para um SKU.
func UpsertStoreAllocation(masterID, lojaID string, qty int, profitShare float64) (*model.StoreAllocation, error) {
	_, err := database.DB.Exec(`
		INSERT INTO store_allocations (id, master_id, loja_id, quantity_allocated, profit_share_pct, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (master_id, loja_id) DO UPDATE
		  SET quantity_allocated = $4, profit_share_pct = $5, updated_at = NOW(), is_active = true
	`, uuid.New(), masterID, lojaID, qty, profitShare)
	if err != nil {
		return nil, fmt.Errorf("inventory: upsert allocation: %w", err)
	}

	row := database.DB.QueryRow(`
		SELECT id, master_id, loja_id, quantity_allocated, quantity_sold,
		       profit_share_pct, is_active, created_at, updated_at
		FROM store_allocations WHERE master_id = $1 AND loja_id = $2
	`, masterID, lojaID)
	return scanAllocationRow(row)
}

// --- InventoryAlerts ---

// ListPendingAlerts lista alertas não confirmados de um tenant.
func ListPendingAlerts(tenantID string) ([]*model.InventoryAlert, error) {
	rows, err := database.DB.Query(`
		SELECT a.id, a.master_id, a.quantity_current, a.quantity_reorder,
		       a.alert_type, a.acknowledged, a.acknowledged_by, a.acknowledged_at, a.created_at
		FROM inventory_alerts a
		JOIN inventory_masters m ON m.id = a.master_id
		WHERE m.tenant_id = $1 AND a.acknowledged = false
		ORDER BY a.created_at DESC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("inventory: list alerts: %w", err)
	}
	defer rows.Close()

	var result []*model.InventoryAlert
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, alert)
	}
	return result, rows.Err()
}

// AcknowledgeAlert marca um alerta como confirmado.
func AcknowledgeAlert(alertID, userID string) error {
	_, err := database.DB.Exec(`
		UPDATE inventory_alerts
		SET acknowledged = true, acknowledged_by = $2, acknowledged_at = NOW()
		WHERE id = $1 AND acknowledged = false
	`, alertID, userID)
	return err
}

// --- InventoryMovements ---

// ListMovements lista as movimentações de um SKU.
func ListMovements(masterID string, limit int) ([]*model.InventoryMovement, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := database.DB.Query(`
		SELECT id, master_id, loja_id, pedido_id, movement_type, quantity,
		       quantity_before, quantity_after, custo_unitario_cents, observacao, created_by, created_at
		FROM inventory_movements
		WHERE master_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, masterID, limit)
	if err != nil {
		return nil, fmt.Errorf("inventory: list movements: %w", err)
	}
	defer rows.Close()

	var result []*model.InventoryMovement
	for rows.Next() {
		m, err := scanMovement(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

// --- SupplierOrders ---

// CreateSupplierOrder cria um pedido de reposição.
func CreateSupplierOrder(tenantID, masterID string, qty int, custoTotal int64, fornecedor, obs string) (*model.SupplierOrder, error) {
	id := uuid.New()
	_, err := database.DB.Exec(`
		INSERT INTO supplier_orders (
			id, tenant_id, master_id, status, quantity_ordered, custo_total_cents,
			fornecedor_nome, observacao, created_at, updated_at
		) VALUES ($1,$2,$3,'rascunho',$4,$5,$6,$7,NOW(),NOW())
	`, id, tenantID, masterID, qty, custoTotal, fornecedor, obs)
	if err != nil {
		return nil, fmt.Errorf("inventory: create supplier order: %w", err)
	}

	row := database.DB.QueryRow(`
		SELECT id, tenant_id, master_id, status, quantity_ordered, quantity_received,
		       custo_total_cents, fornecedor_nome, fornecedor_contato, observacao,
		       expected_at, received_at, created_at, updated_at
		FROM supplier_orders WHERE id = $1
	`, id)
	return scanSupplierOrder(row)
}

// ListSupplierOrders lista pedidos de reposição de um tenant.
func ListSupplierOrders(tenantID string) ([]*model.SupplierOrder, error) {
	rows, err := database.DB.Query(`
		SELECT id, tenant_id, master_id, status, quantity_ordered, quantity_received,
		       custo_total_cents, fornecedor_nome, fornecedor_contato, observacao,
		       expected_at, received_at, created_at, updated_at
		FROM supplier_orders
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("inventory: list orders: %w", err)
	}
	defer rows.Close()

	var result []*model.SupplierOrder
	for rows.Next() {
		o, err := scanSupplierOrderRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, rows.Err()
}

// --- Scanners ---

func scanInventoryMaster(rows *sql.Rows) (*model.InventoryMaster, error) {
	m := &model.InventoryMaster{}
	err := rows.Scan(
		&m.ID, &m.TenantID, &m.ProdutoID, &m.SKUGlobal, &m.Nome, &m.Descricao, &m.Unidade,
		&m.CustoUnitarioCents, &m.QuantityTotal, &m.QuantityReserved,
		&m.ReorderPoint, &m.ReorderQuantity, &m.IsActive, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("inventory: scan master: %w", err)
	}
	return m, nil
}

func scanInventoryMasterRow(row *sql.Row) (*model.InventoryMaster, error) {
	m := &model.InventoryMaster{}
	err := row.Scan(
		&m.ID, &m.TenantID, &m.ProdutoID, &m.SKUGlobal, &m.Nome, &m.Descricao, &m.Unidade,
		&m.CustoUnitarioCents, &m.QuantityTotal, &m.QuantityReserved,
		&m.ReorderPoint, &m.ReorderQuantity, &m.IsActive, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inventory: scan master row: %w", err)
	}
	return m, nil
}

func scanAllocation(rows *sql.Rows) (*model.StoreAllocation, error) {
	a := &model.StoreAllocation{}
	err := rows.Scan(
		&a.ID, &a.MasterID, &a.LojaID, &a.QuantityAllocated, &a.QuantitySold,
		&a.ProfitSharePct, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("inventory: scan allocation: %w", err)
	}
	return a, nil
}

func scanAllocationRow(row *sql.Row) (*model.StoreAllocation, error) {
	a := &model.StoreAllocation{}
	err := row.Scan(
		&a.ID, &a.MasterID, &a.LojaID, &a.QuantityAllocated, &a.QuantitySold,
		&a.ProfitSharePct, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inventory: scan allocation row: %w", err)
	}
	return a, nil
}

func scanAlert(rows *sql.Rows) (*model.InventoryAlert, error) {
	a := &model.InventoryAlert{}
	err := rows.Scan(
		&a.ID, &a.MasterID, &a.QuantityCurrent, &a.QuantityReorder,
		&a.AlertType, &a.Acknowledged, &a.AcknowledgedBy, &a.AcknowledgedAt, &a.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("inventory: scan alert: %w", err)
	}
	return a, nil
}

func scanMovement(rows *sql.Rows) (*model.InventoryMovement, error) {
	m := &model.InventoryMovement{}
	err := rows.Scan(
		&m.ID, &m.MasterID, &m.LojaID, &m.PedidoID, &m.MovementType,
		&m.Quantity, &m.QuantityBefore, &m.QuantityAfter,
		&m.CustoUnitarioCents, &m.Observacao, &m.CreatedBy, &m.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("inventory: scan movement: %w", err)
	}
	return m, nil
}

func scanSupplierOrder(row *sql.Row) (*model.SupplierOrder, error) {
	o := &model.SupplierOrder{}
	err := row.Scan(
		&o.ID, &o.TenantID, &o.MasterID, &o.Status, &o.QuantityOrdered, &o.QuantityReceived,
		&o.CustoTotalCents, &o.FornecedorNome, &o.FornecedorContato, &o.Observacao,
		&o.ExpectedAt, &o.ReceivedAt, &o.CreatedAt, &o.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inventory: scan supplier order: %w", err)
	}
	return o, nil
}

func scanSupplierOrderRows(rows *sql.Rows) (*model.SupplierOrder, error) {
	o := &model.SupplierOrder{}
	err := rows.Scan(
		&o.ID, &o.TenantID, &o.MasterID, &o.Status, &o.QuantityOrdered, &o.QuantityReceived,
		&o.CustoTotalCents, &o.FornecedorNome, &o.FornecedorContato, &o.Observacao,
		&o.ExpectedAt, &o.ReceivedAt, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("inventory: scan supplier order rows: %w", err)
	}
	return o, nil
}
