package model

import "time"

// --- CRM: Clientes ---

type Customer struct {
	ID             string     `json:"id"`
	StoreID        string     `json:"store_id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	Phone          string     `json:"phone"`
	TotalSpentCents int64     `json:"total_spent_cents"`
	LastPurchaseAt *time.Time `json:"last_purchase_at"`
	OrderCount     int        `json:"order_count"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CustomerDetail struct {
	Customer
	Orders []Order `json:"orders"`
}

type CustomerListFilter struct {
	Search  *string
	Page    int
	PerPage int
}

// --- Dashboard ---

type DashboardData struct {
	SalesTodayCents    int64   `json:"sales_today_cents"`
	SalesMonthCents    int64   `json:"sales_month_cents"`
	OrdersPending      int     `json:"orders_pending"`
	NewCustomersMonth  int     `json:"new_customers_month"`
	ConversionRate     float64 `json:"conversion_rate"`
	TotalOrders        int     `json:"total_orders"`
	TotalPaidOrders    int     `json:"total_paid_orders"`
}

// --- Relatórios ---

type SalesReportItem struct {
	Date           string `json:"date"`
	OrderCount     int    `json:"order_count"`
	TotalCents     int64  `json:"total_cents"`
	AvgOrderCents  int64  `json:"avg_order_cents"`
}

type SalesReport struct {
	Items         []SalesReportItem `json:"items"`
	TotalCents    int64             `json:"total_cents"`
	TotalOrders   int               `json:"total_orders"`
	Period        string            `json:"period"`
}

type TopProduct struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	TotalSold   int    `json:"total_sold"`
	RevenueCents int64 `json:"revenue_cents"`
}

type StockAlert struct {
	ProductID          string `json:"product_id"`
	ProductName        string `json:"product_name"`
	SKU                string `json:"sku"`
	StockQuantity      int    `json:"stock_quantity"`
	AlertThreshold     int    `json:"alert_threshold"`
}
