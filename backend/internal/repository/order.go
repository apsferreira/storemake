package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/pkg/database"
)

// CreateOrder cria um pedido com itens dentro de uma transação.
func CreateOrder(order *model.Order, items []model.CartItem) (*model.Order, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	addressJSON, err := json.Marshal(order.Address)
	if err != nil {
		addressJSON = []byte("{}")
	}

	o := &model.Order{}
	var paymentID, notes sql.NullString
	if order.PaymentID != "" {
		paymentID = sql.NullString{String: order.PaymentID, Valid: true}
	}
	if order.Notes != "" {
		notes = sql.NullString{String: order.Notes, Valid: true}
	}

	err = tx.QueryRow(
		`INSERT INTO pedidos (loja_id, customer_name, customer_email, customer_phone,
		 status, subtotal_cents, shipping_cents, discount_cents, total_cents,
		 payment_method, payment_id, shipping_address_json, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		 RETURNING id, loja_id, COALESCE(customer_name, ''), COALESCE(customer_email, ''),
		 COALESCE(customer_phone, ''),
		 status, subtotal_cents, shipping_cents, discount_cents, total_cents,
		 COALESCE(payment_method, ''), COALESCE(payment_id, ''), COALESCE(tracking_code, ''),
		 COALESCE(notes, ''), created_at, updated_at`,
		order.StoreID, order.CustomerName, order.CustomerEmail, order.CustomerPhone,
		"pendente", order.SubtotalCents, order.ShippingCents, order.DiscountCents,
		order.TotalCents, order.PaymentMethod, paymentID, addressJSON, notes,
	).Scan(&o.ID, &o.StoreID, &o.CustomerName, &o.CustomerEmail, &o.CustomerPhone,
		&o.Status, &o.SubtotalCents, &o.ShippingCents, &o.DiscountCents,
		&o.TotalCents, &o.PaymentMethod, &o.PaymentID, &o.TrackingCode,
		&o.Notes, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar pedido: %w", err)
	}

	// Inserir itens do pedido
	for _, ci := range items {
		var variantName string
		if ci.VariantName != "" {
			variantName = ci.VariantName
		}

		oi := model.OrderItem{}
		err = tx.QueryRow(
			`INSERT INTO pedido_itens (pedido_id, produto_id, variacao_id, quantity, unit_price_cents)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, pedido_id, produto_id, COALESCE(variacao_id::text, ''),
			 quantity, unit_price_cents, created_at`,
			o.ID, ci.ProductID, ci.VariantID, ci.Quantity, ci.UnitPriceCents,
		).Scan(&oi.ID, &oi.OrderID, &oi.ProductID, &oi.Variant,
			&oi.Quantity, &oi.UnitPriceCents, &oi.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar item do pedido: %w", err)
		}
		oi.ProductName = ci.ProductName
		if variantName != "" {
			oi.Variant = variantName
		}
		o.Items = append(o.Items, oi)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao commitar transação: %w", err)
	}

	return o, nil
}

// GetOrder retorna um pedido com seus itens.
func GetOrder(storeID, orderID string) (*model.Order, error) {
	o := &model.Order{}
	var addressJSON []byte

	err := database.DB.QueryRow(
		`SELECT id, loja_id, COALESCE(customer_name, ''), COALESCE(customer_email, ''),
		 COALESCE(customer_phone, ''),
		 status, subtotal_cents, shipping_cents, discount_cents, total_cents,
		 COALESCE(payment_method, ''), COALESCE(payment_id, ''),
		 COALESCE(tracking_code, ''), COALESCE(notes, ''),
		 COALESCE(shipping_address_json, '{}'::jsonb), created_at, updated_at
		 FROM pedidos
		 WHERE id = $1 AND loja_id = $2`,
		orderID, storeID,
	).Scan(&o.ID, &o.StoreID, &o.CustomerName, &o.CustomerEmail, &o.CustomerPhone,
		&o.Status, &o.SubtotalCents, &o.ShippingCents, &o.DiscountCents,
		&o.TotalCents, &o.PaymentMethod, &o.PaymentID, &o.TrackingCode, &o.Notes,
		&addressJSON, &o.CreatedAt, &o.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pedido: %w", err)
	}
	if len(addressJSON) > 0 {
		_ = json.Unmarshal(addressJSON, &o.Address)
	}

	// Buscar itens
	items, err := GetOrderItems(o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return o, nil
}

// GetOrderItems retorna os itens de um pedido.
func GetOrderItems(orderID string) ([]model.OrderItem, error) {
	rows, err := database.DB.Query(
		`SELECT pi.id, pi.pedido_id, pi.produto_id,
		        COALESCE(p.name, 'Produto removido') AS product_name,
		        COALESCE(pv.name || ': ' || pv.value, '') AS variant,
		        pi.quantity, pi.unit_price_cents, pi.created_at
		 FROM pedido_itens pi
		 LEFT JOIN produtos p ON p.id = pi.produto_id
		 LEFT JOIN produto_variacoes pv ON pv.id = pi.variacao_id
		 WHERE pi.pedido_id = $1
		 ORDER BY pi.created_at ASC`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar itens do pedido: %w", err)
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID,
			&item.ProductName, &item.Variant, &item.Quantity,
			&item.UnitPriceCents, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("erro ao escanear item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

// ListOrders lista pedidos de uma loja com filtros e paginação.
func ListOrders(storeID string, filter model.OrderListFilter) ([]model.Order, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("loja_id = $%d", argIdx))
	args = append(args, storeID)
	argIdx++

	if filter.Status != nil && *filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(customer_name ILIKE $%d OR customer_email ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM pedidos WHERE %s", where)
	if err := database.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("erro ao contar pedidos: %w", err)
	}

	offset := (filter.Page - 1) * filter.PerPage
	query := fmt.Sprintf(
		`SELECT id, loja_id, COALESCE(customer_name, ''), COALESCE(customer_email, ''),
		 COALESCE(customer_phone, ''),
		 status, subtotal_cents, shipping_cents, discount_cents, total_cents,
		 COALESCE(payment_method, ''), COALESCE(payment_id, ''),
		 COALESCE(tracking_code, ''), COALESCE(notes, ''),
		 created_at, updated_at
		 FROM pedidos
		 WHERE %s
		 ORDER BY created_at DESC
		 LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao listar pedidos: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.StoreID, &o.CustomerName, &o.CustomerEmail,
			&o.CustomerPhone, &o.Status, &o.SubtotalCents, &o.ShippingCents,
			&o.DiscountCents, &o.TotalCents, &o.PaymentMethod, &o.PaymentID,
			&o.TrackingCode, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("erro ao escanear pedido: %w", err)
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

// UpdateOrderStatus atualiza o status de um pedido com validação de transição.
func UpdateOrderStatus(storeID, orderID, newStatus, trackingCode string) (*model.Order, error) {
	// Buscar status atual
	var currentStatus string
	err := database.DB.QueryRow(
		`SELECT status FROM pedidos WHERE id = $1 AND loja_id = $2`,
		orderID, storeID,
	).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("pedido não encontrado")
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pedido: %w", err)
	}

	if !isValidTransition(currentStatus, newStatus) {
		return nil, fmt.Errorf("transição de status inválida: %s → %s", currentStatus, newStatus)
	}

	o := &model.Order{}

	var query string
	var args []interface{}

	returnCols := `RETURNING id, loja_id, COALESCE(customer_name, ''), COALESCE(customer_email, ''),
		         COALESCE(customer_phone, ''), status, subtotal_cents, shipping_cents,
		         discount_cents, total_cents, COALESCE(payment_method, ''),
		         COALESCE(payment_id, ''), COALESCE(tracking_code, ''),
		         COALESCE(notes, ''), created_at, updated_at`

	if trackingCode != "" {
		query = `UPDATE pedidos SET status = $1, tracking_code = $2, updated_at = NOW()
		         WHERE id = $3 AND loja_id = $4 ` + returnCols
		args = []interface{}{newStatus, trackingCode, orderID, storeID}
	} else {
		query = `UPDATE pedidos SET status = $1, updated_at = NOW()
		         WHERE id = $2 AND loja_id = $3 ` + returnCols
		args = []interface{}{newStatus, orderID, storeID}
	}

	err = database.DB.QueryRow(query, args...).Scan(
		&o.ID, &o.StoreID, &o.CustomerName, &o.CustomerEmail,
		&o.CustomerPhone, &o.Status, &o.SubtotalCents, &o.ShippingCents,
		&o.DiscountCents, &o.TotalCents, &o.PaymentMethod, &o.PaymentID,
		&o.TrackingCode, &o.Notes, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar status: %w", err)
	}

	return o, nil
}

// UpdateOrderPayment atualiza o payment_id e status de um pedido (usado pelo webhook).
func UpdateOrderPayment(orderID, paymentID, status string) error {
	result, err := database.DB.Exec(
		`UPDATE pedidos SET payment_id = $1, status = $2, updated_at = NOW()
		 WHERE id = $3`,
		paymentID, status, orderID,
	)
	if err != nil {
		return fmt.Errorf("erro ao atualizar pagamento: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("pedido não encontrado")
	}
	return nil
}

// GetOrderByID busca pedido sem filtrar por loja (usado em webhooks).
func GetOrderByID(orderID string) (*model.Order, error) {
	o := &model.Order{}

	err := database.DB.QueryRow(
		`SELECT id, loja_id, COALESCE(customer_name, ''), COALESCE(customer_email, ''),
		 COALESCE(customer_phone, ''), status, subtotal_cents, shipping_cents, discount_cents,
		 total_cents, COALESCE(payment_method, ''), COALESCE(payment_id, ''),
		 COALESCE(tracking_code, ''), COALESCE(notes, ''),
		 created_at, updated_at
		 FROM pedidos WHERE id = $1`,
		orderID,
	).Scan(&o.ID, &o.StoreID, &o.CustomerName, &o.CustomerEmail,
		&o.CustomerPhone, &o.Status, &o.SubtotalCents, &o.ShippingCents,
		&o.DiscountCents, &o.TotalCents, &o.PaymentMethod, &o.PaymentID,
		&o.TrackingCode, &o.Notes, &o.CreatedAt, &o.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pedido: %w", err)
	}
	return o, nil
}

// GetOrderItemVariantID busca o variacao_id de um item do pedido.
func GetOrderItemVariantID(orderID, productID string, variantID **string) error {
	var vid sql.NullString
	err := database.DB.QueryRow(
		`SELECT variacao_id FROM pedido_itens WHERE pedido_id = $1 AND produto_id = $2 LIMIT 1`,
		orderID, productID,
	).Scan(&vid)
	if err != nil {
		return err
	}
	if vid.Valid {
		*variantID = &vid.String
	}
	return nil
}

// isValidTransition verifica se a transição de status é válida.
func isValidTransition(from, to string) bool {
	validTransitions := map[string][]string{
		"pendente":   {"pago", "cancelado"},
		"pago":       {"preparando", "cancelado"},
		"preparando": {"enviado", "cancelado"},
		"enviado":    {"entregue"},
		"entregue":   {},
		"cancelado":  {},
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// CountPaidOrders retorna quantos pedidos com status "pago" a loja tem.
// Usado para detectar a primeira venda (count == 1 após confirmação).
func CountPaidOrders(storeID string) (int, error) {
	var count int
	err := database.DB.QueryRow(
		`SELECT COUNT(*) FROM pedidos WHERE loja_id = $1 AND status = 'pago'`,
		storeID,
	).Scan(&count)
	return count, err
}
