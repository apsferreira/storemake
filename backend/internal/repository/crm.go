package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/pkg/database"
)

// ListCustomers lista clientes da loja com total_gasto, última compra e num_pedidos.
func ListCustomers(storeID string, filter model.CustomerListFilter) ([]model.Customer, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("c.loja_id = $%d", argIdx))
	args = append(args, storeID)
	argIdx++

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(c.name ILIKE $%d OR c.email ILIKE $%d OR c.phone ILIKE $%d)",
			argIdx, argIdx, argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count
	var total int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM clientes c
		WHERE %s`, where)
	if err := database.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("erro ao contar clientes: %w", err)
	}

	offset := (filter.Page - 1) * filter.PerPage

	query := fmt.Sprintf(`
		SELECT c.id, c.loja_id, c.name, COALESCE(c.email, ''), COALESCE(c.phone, ''),
		       COALESCE(c.total_spent_cents, 0),
		       c.last_purchase_at,
		       COALESCE(cnt.order_count, 0),
		       c.created_at
		FROM clientes c
		LEFT JOIN (
			SELECT cliente_id, COUNT(*) AS order_count
			FROM pedidos
			WHERE loja_id = $1 AND status NOT IN ('cancelado')
			GROUP BY cliente_id
		) cnt ON cnt.cliente_id = c.id
		WHERE %s
		ORDER BY c.created_at DESC
		LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1)
	args = append(args, filter.PerPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao listar clientes: %w", err)
	}
	defer rows.Close()

	var customers []model.Customer
	for rows.Next() {
		var c model.Customer
		var lastPurchase sql.NullTime
		if err := rows.Scan(&c.ID, &c.StoreID, &c.Name, &c.Email, &c.Phone,
			&c.TotalSpentCents, &lastPurchase, &c.OrderCount, &c.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("erro ao escanear cliente: %w", err)
		}
		if lastPurchase.Valid {
			c.LastPurchaseAt = &lastPurchase.Time
		}
		customers = append(customers, c)
	}

	return customers, total, nil
}

// GetCustomerDetail retorna cliente com histórico de pedidos.
func GetCustomerDetail(storeID, customerID string) (*model.CustomerDetail, error) {
	var c model.Customer
	var lastPurchase sql.NullTime

	err := database.DB.QueryRow(`
		SELECT c.id, c.loja_id, c.name, COALESCE(c.email, ''), COALESCE(c.phone, ''),
		       COALESCE(c.total_spent_cents, 0), c.last_purchase_at,
		       COALESCE(cnt.order_count, 0),
		       c.created_at
		FROM clientes c
		LEFT JOIN (
			SELECT cliente_id, COUNT(*) AS order_count
			FROM pedidos
			WHERE loja_id = $2 AND status NOT IN ('cancelado')
			GROUP BY cliente_id
		) cnt ON cnt.cliente_id = c.id
		WHERE c.id = $1 AND c.loja_id = $2`,
		customerID, storeID,
	).Scan(&c.ID, &c.StoreID, &c.Name, &c.Email, &c.Phone,
		&c.TotalSpentCents, &lastPurchase, &c.OrderCount, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar cliente: %w", err)
	}
	if lastPurchase.Valid {
		c.LastPurchaseAt = &lastPurchase.Time
	}

	// Buscar pedidos do cliente
	rows, err := database.DB.Query(`
		SELECT id, loja_id, COALESCE(customer_name, ''), COALESCE(customer_email, ''),
		       COALESCE(customer_phone, ''),
		       status, subtotal_cents, shipping_cents, discount_cents, total_cents,
		       COALESCE(payment_method, ''), COALESCE(payment_id, ''),
		       COALESCE(tracking_code, ''), COALESCE(notes, ''),
		       created_at, updated_at
		FROM pedidos
		WHERE cliente_id = $1 AND loja_id = $2
		ORDER BY created_at DESC
		LIMIT 50`,
		customerID, storeID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pedidos do cliente: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.StoreID, &o.CustomerName, &o.CustomerEmail,
			&o.CustomerPhone, &o.Status, &o.SubtotalCents, &o.ShippingCents,
			&o.DiscountCents, &o.TotalCents, &o.PaymentMethod, &o.PaymentID,
			&o.TrackingCode, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("erro ao escanear pedido: %w", err)
		}
		orders = append(orders, o)
	}

	return &model.CustomerDetail{
		Customer: c,
		Orders:   orders,
	}, nil
}

// GetDashboard retorna dados do dashboard admin.
func GetDashboard(storeID string) (*model.DashboardData, error) {
	d := &model.DashboardData{}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Vendas hoje (apenas pedidos pagos/preparando/enviado/entregue)
	err := database.DB.QueryRow(`
		SELECT COALESCE(SUM(total_cents), 0)
		FROM pedidos
		WHERE loja_id = $1
		  AND status NOT IN ('pendente', 'cancelado')
		  AND created_at >= $2`,
		storeID, todayStart,
	).Scan(&d.SalesTodayCents)
	if err != nil {
		return nil, fmt.Errorf("erro ao calcular vendas hoje: %w", err)
	}

	// Vendas do mês
	err = database.DB.QueryRow(`
		SELECT COALESCE(SUM(total_cents), 0)
		FROM pedidos
		WHERE loja_id = $1
		  AND status NOT IN ('pendente', 'cancelado')
		  AND created_at >= $2`,
		storeID, monthStart,
	).Scan(&d.SalesMonthCents)
	if err != nil {
		return nil, fmt.Errorf("erro ao calcular vendas mês: %w", err)
	}

	// Pedidos pendentes
	err = database.DB.QueryRow(`
		SELECT COUNT(*)
		FROM pedidos
		WHERE loja_id = $1 AND status = 'pendente'`,
		storeID,
	).Scan(&d.OrdersPending)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar pedidos pendentes: %w", err)
	}

	// Novos clientes no mês
	err = database.DB.QueryRow(`
		SELECT COUNT(*)
		FROM clientes
		WHERE loja_id = $1 AND created_at >= $2`,
		storeID, monthStart,
	).Scan(&d.NewCustomersMonth)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar novos clientes: %w", err)
	}

	// Taxa de conversão: pedidos pagos+ / total pedidos
	err = database.DB.QueryRow(`
		SELECT
			COALESCE(COUNT(*), 0) AS total,
			COALESCE(SUM(CASE WHEN status NOT IN ('pendente', 'cancelado') THEN 1 ELSE 0 END), 0) AS paid
		FROM pedidos
		WHERE loja_id = $1 AND created_at >= $2`,
		storeID, monthStart,
	).Scan(&d.TotalOrders, &d.TotalPaidOrders)
	if err != nil {
		return nil, fmt.Errorf("erro ao calcular conversão: %w", err)
	}

	if d.TotalOrders > 0 {
		d.ConversionRate = float64(d.TotalPaidOrders) / float64(d.TotalOrders) * 100
	}

	return d, nil
}

// GetSalesReport retorna relatório de vendas por período agrupado por dia.
func GetSalesReport(storeID, from, to string) (*model.SalesReport, error) {
	rows, err := database.DB.Query(`
		SELECT
			TO_CHAR(created_at, 'YYYY-MM-DD') AS sale_date,
			COUNT(*) AS order_count,
			COALESCE(SUM(total_cents), 0) AS total_cents,
			COALESCE(AVG(total_cents), 0)::BIGINT AS avg_order_cents
		FROM pedidos
		WHERE loja_id = $1
		  AND status NOT IN ('cancelado')
		  AND created_at >= $2::DATE
		  AND created_at < ($3::DATE + INTERVAL '1 day')
		GROUP BY sale_date
		ORDER BY sale_date ASC`,
		storeID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar relatório de vendas: %w", err)
	}
	defer rows.Close()

	report := &model.SalesReport{
		Period: fmt.Sprintf("%s a %s", from, to),
	}

	for rows.Next() {
		var item model.SalesReportItem
		if err := rows.Scan(&item.Date, &item.OrderCount, &item.TotalCents, &item.AvgOrderCents); err != nil {
			return nil, fmt.Errorf("erro ao escanear relatório: %w", err)
		}
		report.Items = append(report.Items, item)
		report.TotalCents += item.TotalCents
		report.TotalOrders += item.OrderCount
	}

	if report.Items == nil {
		report.Items = []model.SalesReportItem{}
	}

	return report, nil
}

// GetTopProducts retorna top 5 produtos por vendas e receita.
func GetTopProducts(storeID string) ([]model.TopProduct, error) {
	rows, err := database.DB.Query(`
		SELECT
			pi.produto_id,
			COALESCE(p.name, 'Produto removido') AS product_name,
			COALESCE(SUM(pi.quantity), 0) AS total_sold,
			COALESCE(SUM(pi.quantity * pi.unit_price_cents), 0) AS revenue_cents
		FROM pedido_itens pi
		JOIN pedidos ped ON ped.id = pi.pedido_id
		LEFT JOIN produtos p ON p.id = pi.produto_id
		WHERE ped.loja_id = $1
		  AND ped.status NOT IN ('cancelado')
		GROUP BY pi.produto_id, p.name
		ORDER BY revenue_cents DESC
		LIMIT 5`,
		storeID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar top produtos: %w", err)
	}
	defer rows.Close()

	var products []model.TopProduct
	for rows.Next() {
		var tp model.TopProduct
		if err := rows.Scan(&tp.ProductID, &tp.ProductName, &tp.TotalSold, &tp.RevenueCents); err != nil {
			return nil, fmt.Errorf("erro ao escanear top produto: %w", err)
		}
		products = append(products, tp)
	}

	if products == nil {
		products = []model.TopProduct{}
	}

	return products, nil
}

// GetStockAlerts retorna produtos com estoque abaixo do threshold.
func GetStockAlerts(storeID string) ([]model.StockAlert, error) {
	rows, err := database.DB.Query(`
		SELECT id, name, COALESCE(sku, ''), stock_quantity, stock_alert_threshold
		FROM produtos
		WHERE loja_id = $1
		  AND is_active = true
		  AND stock_quantity <= stock_alert_threshold
		ORDER BY stock_quantity ASC`,
		storeID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar alertas de estoque: %w", err)
	}
	defer rows.Close()

	var alerts []model.StockAlert
	for rows.Next() {
		var a model.StockAlert
		if err := rows.Scan(&a.ProductID, &a.ProductName, &a.SKU, &a.StockQuantity, &a.AlertThreshold); err != nil {
			return nil, fmt.Errorf("erro ao escanear alerta: %w", err)
		}
		alerts = append(alerts, a)
	}

	if alerts == nil {
		alerts = []model.StockAlert{}
	}

	return alerts, nil
}

// GetSalesForExport retorna dados de vendas para export CSV.
func GetSalesForExport(storeID, from, to string) ([][]string, error) {
	rows, err := database.DB.Query(`
		SELECT
			ped.id,
			TO_CHAR(ped.created_at, 'YYYY-MM-DD HH24:MI') AS data,
			COALESCE(ped.customer_name, '') AS cliente,
			COALESCE(ped.customer_email, '') AS email,
			ped.status,
			ped.subtotal_cents,
			ped.shipping_cents,
			ped.discount_cents,
			ped.total_cents,
			COALESCE(ped.payment_method, '') AS metodo
		FROM pedidos ped
		WHERE ped.loja_id = $1
		  AND ped.status NOT IN ('cancelado')
		  AND ped.created_at >= $2::DATE
		  AND ped.created_at < ($3::DATE + INTERVAL '1 day')
		ORDER BY ped.created_at ASC`,
		storeID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar vendas para export: %w", err)
	}
	defer rows.Close()

	var records [][]string
	// Header
	records = append(records, []string{
		"ID", "Data", "Cliente", "Email", "Status",
		"Subtotal (centavos)", "Frete (centavos)", "Desconto (centavos)",
		"Total (centavos)", "Método Pagamento",
	})

	for rows.Next() {
		var id, data, cliente, email, status, metodo string
		var subtotal, shipping, discount, total int64
		if err := rows.Scan(&id, &data, &cliente, &email, &status,
			&subtotal, &shipping, &discount, &total, &metodo); err != nil {
			return nil, fmt.Errorf("erro ao escanear venda: %w", err)
		}
		records = append(records, []string{
			id, data, cliente, email, status,
			fmt.Sprintf("%d", subtotal),
			fmt.Sprintf("%d", shipping),
			fmt.Sprintf("%d", discount),
			fmt.Sprintf("%d", total),
			metodo,
		})
	}

	return records, nil
}
