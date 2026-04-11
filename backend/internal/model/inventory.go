package model

import (
	"time"

	"github.com/google/uuid"
)

// InventoryMaster representa o SKU centralizado de um tenant multi-loja. BKL-900.
type InventoryMaster struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	TenantID           uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	ProdutoID          *uuid.UUID `json:"produto_id,omitempty" db:"produto_id"`
	SKUGlobal          string     `json:"sku_global" db:"sku_global"`
	Nome               string     `json:"nome" db:"nome"`
	Descricao          string     `json:"descricao,omitempty" db:"descricao"`
	Unidade            string     `json:"unidade" db:"unidade"`
	CustoUnitarioCents int64      `json:"custo_unitario_cents" db:"custo_unitario_cents"`
	QuantityTotal      int        `json:"quantity_total" db:"quantity_total"`
	QuantityReserved   int        `json:"quantity_reserved" db:"quantity_reserved"`
	ReorderPoint       int        `json:"reorder_point" db:"reorder_point"`
	ReorderQuantity    int        `json:"reorder_quantity" db:"reorder_quantity"`
	IsActive           bool       `json:"is_active" db:"is_active"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// QuantityAvailable retorna a quantidade disponível (total - reservada - alocada).
func (m *InventoryMaster) QuantityAvailable() int {
	available := m.QuantityTotal - m.QuantityReserved
	if available < 0 {
		return 0
	}
	return available
}

// StoreAllocation representa a cota de estoque alocada para uma loja específica. BKL-900.
type StoreAllocation struct {
	ID               uuid.UUID `json:"id" db:"id"`
	MasterID         uuid.UUID `json:"master_id" db:"master_id"`
	LojaID           uuid.UUID `json:"loja_id" db:"loja_id"`
	QuantityAllocated int      `json:"quantity_allocated" db:"quantity_allocated"`
	QuantitySold     int       `json:"quantity_sold" db:"quantity_sold"`
	ProfitSharePct   float64   `json:"profit_share_pct" db:"profit_share_pct"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// MovementType define os tipos de movimentação de estoque.
type MovementType string

const (
	MovementTypeEntrada     MovementType = "entrada"
	MovementTypeSaidaVenda  MovementType = "saida_venda"
	MovementTypeSaidaPerda  MovementType = "saida_perda"
	MovementTypeTransferencia MovementType = "transferencia"
	MovementTypeAjuste      MovementType = "ajuste"
	MovementTypeDevolucao   MovementType = "devolucao"
)

// InventoryMovement registra cada movimentação de estoque para auditoria. BKL-900.
type InventoryMovement struct {
	ID                  uuid.UUID    `json:"id" db:"id"`
	MasterID            uuid.UUID    `json:"master_id" db:"master_id"`
	LojaID              *uuid.UUID   `json:"loja_id,omitempty" db:"loja_id"`
	PedidoID            *uuid.UUID   `json:"pedido_id,omitempty" db:"pedido_id"`
	MovementType        MovementType `json:"movement_type" db:"movement_type"`
	Quantity            int          `json:"quantity" db:"quantity"`
	QuantityBefore      int          `json:"quantity_before" db:"quantity_before"`
	QuantityAfter       int          `json:"quantity_after" db:"quantity_after"`
	CustoUnitarioCents  *int64       `json:"custo_unitario_cents,omitempty" db:"custo_unitario_cents"`
	Observacao          string       `json:"observacao,omitempty" db:"observacao"`
	CreatedBy           *uuid.UUID   `json:"created_by,omitempty" db:"created_by"`
	CreatedAt           time.Time    `json:"created_at" db:"created_at"`
}

// InventoryAlert é gerado automaticamente quando quantity_total <= reorder_point. BKL-900.
type InventoryAlert struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	MasterID        uuid.UUID  `json:"master_id" db:"master_id"`
	QuantityCurrent int        `json:"quantity_current" db:"quantity_current"`
	QuantityReorder int        `json:"quantity_reorder" db:"quantity_reorder"`
	AlertType       string     `json:"alert_type" db:"alert_type"` // low_stock | out_of_stock
	Acknowledged    bool       `json:"acknowledged" db:"acknowledged"`
	AcknowledgedBy  *uuid.UUID `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// SupplierOrderStatus define os estados de um pedido de reposição.
type SupplierOrderStatus string

const (
	SupplierOrderRascunho  SupplierOrderStatus = "rascunho"
	SupplierOrderEnviado   SupplierOrderStatus = "enviado"
	SupplierOrderConfirmado SupplierOrderStatus = "confirmado"
	SupplierOrderRecebido  SupplierOrderStatus = "recebido"
	SupplierOrderCancelado SupplierOrderStatus = "cancelado"
)

// SupplierOrder representa um pedido de reposição ao fornecedor. BKL-900.
type SupplierOrder struct {
	ID               uuid.UUID           `json:"id" db:"id"`
	TenantID         uuid.UUID           `json:"tenant_id" db:"tenant_id"`
	MasterID         uuid.UUID           `json:"master_id" db:"master_id"`
	Status           SupplierOrderStatus `json:"status" db:"status"`
	QuantityOrdered  int                 `json:"quantity_ordered" db:"quantity_ordered"`
	QuantityReceived int                 `json:"quantity_received" db:"quantity_received"`
	CustoTotalCents  int64               `json:"custo_total_cents" db:"custo_total_cents"`
	FornecedorNome   string              `json:"fornecedor_nome,omitempty" db:"fornecedor_nome"`
	FornecedorContato string             `json:"fornecedor_contato,omitempty" db:"fornecedor_contato"`
	Observacao       string              `json:"observacao,omitempty" db:"observacao"`
	ExpectedAt       *time.Time          `json:"expected_at,omitempty" db:"expected_at"`
	ReceivedAt       *time.Time          `json:"received_at,omitempty" db:"received_at"`
	CreatedAt        time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at" db:"updated_at"`
}

// --- Request types ---

// CreateInventoryMasterRequest é usado para criar um novo SKU centralizado.
type CreateInventoryMasterRequest struct {
	TenantID           uuid.UUID  `json:"tenant_id"`
	ProdutoID          *uuid.UUID `json:"produto_id,omitempty"`
	SKUGlobal          string     `json:"sku_global"`
	Nome               string     `json:"nome"`
	Descricao          string     `json:"descricao,omitempty"`
	Unidade            string     `json:"unidade"`
	CustoUnitarioCents int64      `json:"custo_unitario_cents"`
	QuantityTotal      int        `json:"quantity_total"`
	ReorderPoint       int        `json:"reorder_point"`
	ReorderQuantity    int        `json:"reorder_quantity"`
}

// UpsertAllocationRequest é usado para alocar/atualizar cota de uma loja.
type UpsertAllocationRequest struct {
	LojaID            string  `json:"loja_id"`
	QuantityAllocated int     `json:"quantity_allocated"`
	ProfitSharePct    float64 `json:"profit_share_pct"`
}

// AdjustQuantityRequest é usado para ajuste manual de estoque.
type AdjustQuantityRequest struct {
	Delta      int    `json:"delta"`       // positivo = entrada, negativo = saída
	LojaID     string `json:"loja_id,omitempty"`
	Observacao string `json:"observacao,omitempty"`
}

// CreateSupplierOrderRequest é usado para criar pedido de reposição.
type CreateSupplierOrderRequest struct {
	QuantityOrdered   int    `json:"quantity_ordered"`
	CustoTotalCents   int64  `json:"custo_total_cents"`
	FornecedorNome    string `json:"fornecedor_nome,omitempty"`
	Observacao        string `json:"observacao,omitempty"`
}

// UpdateInventoryMasterRequest é usado para atualizar os campos cadastrais de um SKU. BKL-900.
type UpdateInventoryMasterRequest struct {
	SKUGlobal          string `json:"sku_global"`
	Nome               string `json:"nome"            validate:"required,max=255"`
	Descricao          string `json:"descricao,omitempty"`
	Unidade            string `json:"unidade"         validate:"required,max=20"`
	CustoUnitarioCents int64  `json:"custo_unitario_cents"`
	ReorderPoint       int    `json:"reorder_point"`
	ReorderQuantity    int    `json:"reorder_quantity"`
}

// RegisterMovementRequest é usado para registrar movimentação explícita (in/out/ajuste). BKL-900.
type RegisterMovementRequest struct {
	MovementType MovementType `json:"movement_type" validate:"required"`
	Delta        int          `json:"delta"`         // positivo = entrada, negativo = saída
	LojaID       string       `json:"loja_id,omitempty"`
	Observacao   string       `json:"observacao,omitempty"`
}
