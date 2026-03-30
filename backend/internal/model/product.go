package model

import "time"

type Product struct {
	ID                  string             `json:"id"`
	LojaID              string             `json:"loja_id"`
	CategoriaID         *string            `json:"categoria_id"`
	Name                string             `json:"name"`
	Slug                string             `json:"slug"`
	Description         *string            `json:"description"`
	PriceCents          int64              `json:"price_cents"`
	ComparePriceCents   int64              `json:"compare_price_cents"`
	SKU                 *string            `json:"sku"`
	StockQuantity       int                `json:"stock_quantity"`
	StockAlertThreshold int                `json:"stock_alert_threshold"`
	IsActive            bool               `json:"is_active"`
	SortOrder           int                `json:"sort_order"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
	Variations          []ProductVariation `json:"variations,omitempty"`
	Photos              []ProductPhoto     `json:"photos,omitempty"`
}

type ProductPhoto struct {
	ID        string    `json:"id"`
	ProdutoID string    `json:"produto_id"`
	URL       string    `json:"url"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductVariation struct {
	ID                   string    `json:"id"`
	ProdutoID            string    `json:"produto_id"`
	Name                 string    `json:"name"`
	Value                string    `json:"value"`
	PriceAdjustmentCents int64     `json:"price_adjustment_cents"`
	StockQuantity        int       `json:"stock_quantity"`
	CreatedAt            time.Time `json:"created_at"`
}

type CreateProductRequest struct {
	CategoriaID         *string                      `json:"categoria_id"`
	Name                string                       `json:"name"`
	Slug                string                       `json:"slug"`
	Description         *string                      `json:"description"`
	PriceCents          int64                        `json:"price_cents"`
	ComparePriceCents   int64                        `json:"compare_price_cents"`
	SKU                 *string                      `json:"sku"`
	StockQuantity       int                          `json:"stock_quantity"`
	StockAlertThreshold int                          `json:"stock_alert_threshold"`
	IsActive            *bool                        `json:"is_active"`
	SortOrder           int                          `json:"sort_order"`
	Variations          []CreateVariationRequest     `json:"variations,omitempty"`
}

type UpdateProductRequest struct {
	CategoriaID         *string `json:"categoria_id,omitempty"`
	Name                *string `json:"name,omitempty"`
	Slug                *string `json:"slug,omitempty"`
	Description         *string `json:"description,omitempty"`
	PriceCents          *int64  `json:"price_cents,omitempty"`
	ComparePriceCents   *int64  `json:"compare_price_cents,omitempty"`
	SKU                 *string `json:"sku,omitempty"`
	StockQuantity       *int    `json:"stock_quantity,omitempty"`
	StockAlertThreshold *int    `json:"stock_alert_threshold,omitempty"`
	IsActive            *bool   `json:"is_active,omitempty"`
	SortOrder           *int    `json:"sort_order,omitempty"`
}

type CreateVariationRequest struct {
	Name                 string `json:"name"`
	Value                string `json:"value"`
	PriceAdjustmentCents int64  `json:"price_adjustment_cents"`
	StockQuantity        int    `json:"stock_quantity"`
}

type ReorderItem struct {
	ID        string `json:"id"`
	SortOrder int    `json:"sort_order"`
}

type ReorderRequest struct {
	Items []ReorderItem `json:"items"`
}

type ProductListFilter struct {
	CategoriaID *string
	MinPrice    *int64
	MaxPrice    *int64
	InStock     *bool
	IsActive    *bool
	Search      *string
	Page        int
	PerPage     int
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}
