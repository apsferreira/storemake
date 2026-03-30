package repository

import (
	"database/sql"
	"fmt"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/pkg/database"
)

func CreateCategory(lojaID string, req model.CreateCategoryRequest) (*model.Category, error) {
	cat := &model.Category{}
	err := database.DB.QueryRow(
		`INSERT INTO categorias (loja_id, name, slug, sort_order)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, loja_id, name, slug, sort_order, created_at`,
		lojaID, req.Name, req.Slug, req.SortOrder,
	).Scan(&cat.ID, &cat.LojaID, &cat.Name, &cat.Slug, &cat.SortOrder, &cat.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar categoria: %w", err)
	}
	return cat, nil
}

func ListCategories(lojaID string) ([]model.Category, error) {
	rows, err := database.DB.Query(
		`SELECT id, loja_id, name, slug, sort_order, created_at
		 FROM categorias
		 WHERE loja_id = $1
		 ORDER BY sort_order ASC, name ASC`,
		lojaID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar categorias: %w", err)
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.LojaID, &c.Name, &c.Slug, &c.SortOrder, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("erro ao escanear categoria: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func GetCategory(lojaID, categoryID string) (*model.Category, error) {
	cat := &model.Category{}
	err := database.DB.QueryRow(
		`SELECT id, loja_id, name, slug, sort_order, created_at
		 FROM categorias
		 WHERE id = $1 AND loja_id = $2`,
		categoryID, lojaID,
	).Scan(&cat.ID, &cat.LojaID, &cat.Name, &cat.Slug, &cat.SortOrder, &cat.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar categoria: %w", err)
	}
	return cat, nil
}

func UpdateCategory(lojaID, categoryID string, req model.UpdateCategoryRequest) (*model.Category, error) {
	cat := &model.Category{}
	err := database.DB.QueryRow(
		`UPDATE categorias
		 SET name = COALESCE($1, name),
		     slug = COALESCE($2, slug),
		     sort_order = COALESCE($3, sort_order)
		 WHERE id = $4 AND loja_id = $5
		 RETURNING id, loja_id, name, slug, sort_order, created_at`,
		req.Name, req.Slug, req.SortOrder, categoryID, lojaID,
	).Scan(&cat.ID, &cat.LojaID, &cat.Name, &cat.Slug, &cat.SortOrder, &cat.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar categoria: %w", err)
	}
	return cat, nil
}

func DeleteCategory(lojaID, categoryID string) error {
	result, err := database.DB.Exec(
		`DELETE FROM categorias WHERE id = $1 AND loja_id = $2`,
		categoryID, lojaID,
	)
	if err != nil {
		return fmt.Errorf("erro ao deletar categoria: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("categoria não encontrada")
	}
	return nil
}
