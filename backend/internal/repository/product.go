package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/pkg/database"
)

func CreateProduct(lojaID string, req model.CreateProductRequest) (*model.Product, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	p := &model.Product{}
	err = tx.QueryRow(
		`INSERT INTO produtos (loja_id, categoria_id, name, slug, description, price_cents,
		 compare_price_cents, sku, stock_quantity, stock_alert_threshold, is_active, sort_order)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING id, loja_id, categoria_id, name, slug, description, price_cents,
		 compare_price_cents, sku, stock_quantity, stock_alert_threshold, is_active,
		 sort_order, created_at, updated_at`,
		lojaID, req.CategoriaID, req.Name, req.Slug, req.Description, req.PriceCents,
		req.ComparePriceCents, req.SKU, req.StockQuantity, req.StockAlertThreshold,
		isActive, req.SortOrder,
	).Scan(&p.ID, &p.LojaID, &p.CategoriaID, &p.Name, &p.Slug, &p.Description,
		&p.PriceCents, &p.ComparePriceCents, &p.SKU, &p.StockQuantity,
		&p.StockAlertThreshold, &p.IsActive, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar produto: %w", err)
	}

	for _, v := range req.Variations {
		var variation model.ProductVariation
		err = tx.QueryRow(
			`INSERT INTO produto_variacoes (produto_id, name, value, price_adjustment_cents, stock_quantity)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, produto_id, name, value, price_adjustment_cents, stock_quantity, created_at`,
			p.ID, v.Name, v.Value, v.PriceAdjustmentCents, v.StockQuantity,
		).Scan(&variation.ID, &variation.ProdutoID, &variation.Name, &variation.Value,
			&variation.PriceAdjustmentCents, &variation.StockQuantity, &variation.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar variação: %w", err)
		}
		p.Variations = append(p.Variations, variation)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao commitar transação: %w", err)
	}

	return p, nil
}

func GetProduct(lojaID, productID string) (*model.Product, error) {
	p := &model.Product{}
	err := database.DB.QueryRow(
		`SELECT id, loja_id, categoria_id, name, slug, description, price_cents,
		 compare_price_cents, sku, stock_quantity, stock_alert_threshold, is_active,
		 sort_order, created_at, updated_at
		 FROM produtos
		 WHERE id = $1 AND loja_id = $2`,
		productID, lojaID,
	).Scan(&p.ID, &p.LojaID, &p.CategoriaID, &p.Name, &p.Slug, &p.Description,
		&p.PriceCents, &p.ComparePriceCents, &p.SKU, &p.StockQuantity,
		&p.StockAlertThreshold, &p.IsActive, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar produto: %w", err)
	}

	variations, err := GetProductVariations(productID)
	if err != nil {
		return nil, err
	}
	p.Variations = variations

	photos, err := GetProductPhotos(productID)
	if err != nil {
		return nil, err
	}
	p.Photos = photos

	return p, nil
}

func ListProducts(lojaID string, filter model.ProductListFilter) ([]model.Product, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("loja_id = $%d", argIdx))
	args = append(args, lojaID)
	argIdx++

	if filter.CategoriaID != nil {
		conditions = append(conditions, fmt.Sprintf("categoria_id = $%d", argIdx))
		args = append(args, *filter.CategoriaID)
		argIdx++
	}
	if filter.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price_cents >= $%d", argIdx))
		args = append(args, *filter.MinPrice)
		argIdx++
	}
	if filter.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price_cents <= $%d", argIdx))
		args = append(args, *filter.MaxPrice)
		argIdx++
	}
	if filter.InStock != nil && *filter.InStock {
		conditions = append(conditions, "stock_quantity > 0")
	}
	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}
	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM produtos WHERE %s", where)
	if err := database.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("erro ao contar produtos: %w", err)
	}

	// Paginated query
	offset := (filter.Page - 1) * filter.PerPage
	query := fmt.Sprintf(
		`SELECT id, loja_id, categoria_id, name, slug, description, price_cents,
		 compare_price_cents, sku, stock_quantity, stock_alert_threshold, is_active,
		 sort_order, created_at, updated_at
		 FROM produtos
		 WHERE %s
		 ORDER BY sort_order ASC, created_at DESC
		 LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao listar produtos: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.LojaID, &p.CategoriaID, &p.Name, &p.Slug,
			&p.Description, &p.PriceCents, &p.ComparePriceCents, &p.SKU,
			&p.StockQuantity, &p.StockAlertThreshold, &p.IsActive,
			&p.SortOrder, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("erro ao escanear produto: %w", err)
		}
		products = append(products, p)
	}

	return products, total, nil
}

func UpdateProduct(lojaID, productID string, req model.UpdateProductRequest) (*model.Product, error) {
	p := &model.Product{}
	err := database.DB.QueryRow(
		`UPDATE produtos SET
		 categoria_id = COALESCE($1, categoria_id),
		 name = COALESCE($2, name),
		 slug = COALESCE($3, slug),
		 description = COALESCE($4, description),
		 price_cents = COALESCE($5, price_cents),
		 compare_price_cents = COALESCE($6, compare_price_cents),
		 sku = COALESCE($7, sku),
		 stock_quantity = COALESCE($8, stock_quantity),
		 stock_alert_threshold = COALESCE($9, stock_alert_threshold),
		 is_active = COALESCE($10, is_active),
		 sort_order = COALESCE($11, sort_order),
		 updated_at = NOW()
		 WHERE id = $12 AND loja_id = $13
		 RETURNING id, loja_id, categoria_id, name, slug, description, price_cents,
		 compare_price_cents, sku, stock_quantity, stock_alert_threshold, is_active,
		 sort_order, created_at, updated_at`,
		req.CategoriaID, req.Name, req.Slug, req.Description, req.PriceCents,
		req.ComparePriceCents, req.SKU, req.StockQuantity, req.StockAlertThreshold,
		req.IsActive, req.SortOrder, productID, lojaID,
	).Scan(&p.ID, &p.LojaID, &p.CategoriaID, &p.Name, &p.Slug, &p.Description,
		&p.PriceCents, &p.ComparePriceCents, &p.SKU, &p.StockQuantity,
		&p.StockAlertThreshold, &p.IsActive, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar produto: %w", err)
	}
	return p, nil
}

func DeleteProduct(lojaID, productID string) error {
	result, err := database.DB.Exec(
		`DELETE FROM produtos WHERE id = $1 AND loja_id = $2`,
		productID, lojaID,
	)
	if err != nil {
		return fmt.Errorf("erro ao deletar produto: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("produto não encontrado")
	}
	return nil
}

func ReorderProducts(lojaID string, items []model.ReorderItem) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	for _, item := range items {
		result, err := tx.Exec(
			`UPDATE produtos SET sort_order = $1, updated_at = NOW()
			 WHERE id = $2 AND loja_id = $3`,
			item.SortOrder, item.ID, lojaID,
		)
		if err != nil {
			return fmt.Errorf("erro ao reordenar produto %s: %w", item.ID, err)
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("produto %s não encontrado na loja", item.ID)
		}
	}

	return tx.Commit()
}

func DeductStock(productID string, quantity int, variationID *string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	// Deduct from product
	var currentStock int
	err = tx.QueryRow(
		`UPDATE produtos SET stock_quantity = stock_quantity - $1, updated_at = NOW()
		 WHERE id = $2 AND stock_quantity >= $1
		 RETURNING stock_quantity`,
		quantity, productID,
	).Scan(&currentStock)
	if err == sql.ErrNoRows {
		return fmt.Errorf("estoque insuficiente para o produto")
	}
	if err != nil {
		return fmt.Errorf("erro ao baixar estoque: %w", err)
	}

	// Deduct from variation if applicable
	if variationID != nil {
		_, err = tx.Exec(
			`UPDATE produto_variacoes SET stock_quantity = stock_quantity - $1
			 WHERE id = $2 AND stock_quantity >= $1`,
			quantity, *variationID,
		)
		if err != nil {
			return fmt.Errorf("erro ao baixar estoque da variação: %w", err)
		}
	}

	return tx.Commit()
}

func GetLowStockProducts(lojaID string) ([]model.Product, error) {
	rows, err := database.DB.Query(
		`SELECT id, loja_id, categoria_id, name, slug, description, price_cents,
		 compare_price_cents, sku, stock_quantity, stock_alert_threshold, is_active,
		 sort_order, created_at, updated_at
		 FROM produtos
		 WHERE loja_id = $1 AND stock_quantity <= stock_alert_threshold AND is_active = true
		 ORDER BY stock_quantity ASC`,
		lojaID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar produtos com estoque baixo: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.LojaID, &p.CategoriaID, &p.Name, &p.Slug,
			&p.Description, &p.PriceCents, &p.ComparePriceCents, &p.SKU,
			&p.StockQuantity, &p.StockAlertThreshold, &p.IsActive,
			&p.SortOrder, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("erro ao escanear produto: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

// Public catalog query - no auth needed, only active products
func ListPublicCatalog(storeID string, filter model.ProductListFilter) ([]model.Product, int, error) {
	active := true
	filter.IsActive = &active
	return ListProducts(storeID, filter)
}
