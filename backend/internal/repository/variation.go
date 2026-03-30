package repository

import (
	"database/sql"
	"fmt"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/pkg/database"
)

func GetProductVariations(productID string) ([]model.ProductVariation, error) {
	rows, err := database.DB.Query(
		`SELECT id, produto_id, name, value, price_adjustment_cents, stock_quantity, created_at
		 FROM produto_variacoes
		 WHERE produto_id = $1
		 ORDER BY name ASC, value ASC`,
		productID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar variações: %w", err)
	}
	defer rows.Close()

	var variations []model.ProductVariation
	for rows.Next() {
		var v model.ProductVariation
		if err := rows.Scan(&v.ID, &v.ProdutoID, &v.Name, &v.Value,
			&v.PriceAdjustmentCents, &v.StockQuantity, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("erro ao escanear variação: %w", err)
		}
		variations = append(variations, v)
	}
	return variations, nil
}

func CreateVariation(productID string, req model.CreateVariationRequest) (*model.ProductVariation, error) {
	v := &model.ProductVariation{}
	err := database.DB.QueryRow(
		`INSERT INTO produto_variacoes (produto_id, name, value, price_adjustment_cents, stock_quantity)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, produto_id, name, value, price_adjustment_cents, stock_quantity, created_at`,
		productID, req.Name, req.Value, req.PriceAdjustmentCents, req.StockQuantity,
	).Scan(&v.ID, &v.ProdutoID, &v.Name, &v.Value,
		&v.PriceAdjustmentCents, &v.StockQuantity, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar variação: %w", err)
	}
	return v, nil
}

func DeleteVariation(variationID string) error {
	result, err := database.DB.Exec(
		`DELETE FROM produto_variacoes WHERE id = $1`,
		variationID,
	)
	if err != nil {
		return fmt.Errorf("erro ao deletar variação: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("variação não encontrada")
	}
	return nil
}

func GetVariation(variationID string) (*model.ProductVariation, error) {
	v := &model.ProductVariation{}
	err := database.DB.QueryRow(
		`SELECT id, produto_id, name, value, price_adjustment_cents, stock_quantity, created_at
		 FROM produto_variacoes
		 WHERE id = $1`,
		variationID,
	).Scan(&v.ID, &v.ProdutoID, &v.Name, &v.Value,
		&v.PriceAdjustmentCents, &v.StockQuantity, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar variação: %w", err)
	}
	return v, nil
}
