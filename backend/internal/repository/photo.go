package repository

import (
	"fmt"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/pkg/database"
)

func GetProductPhotos(productID string) ([]model.ProductPhoto, error) {
	rows, err := database.DB.Query(
		`SELECT id, produto_id, url, sort_order, created_at
		 FROM produto_fotos
		 WHERE produto_id = $1
		 ORDER BY sort_order ASC`,
		productID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar fotos: %w", err)
	}
	defer rows.Close()

	var photos []model.ProductPhoto
	for rows.Next() {
		var ph model.ProductPhoto
		if err := rows.Scan(&ph.ID, &ph.ProdutoID, &ph.URL, &ph.SortOrder, &ph.CreatedAt); err != nil {
			return nil, fmt.Errorf("erro ao escanear foto: %w", err)
		}
		photos = append(photos, ph)
	}
	return photos, nil
}

func CreatePhoto(productID, url string, sortOrder int) (*model.ProductPhoto, error) {
	ph := &model.ProductPhoto{}
	err := database.DB.QueryRow(
		`INSERT INTO produto_fotos (produto_id, url, sort_order)
		 VALUES ($1, $2, $3)
		 RETURNING id, produto_id, url, sort_order, created_at`,
		productID, url, sortOrder,
	).Scan(&ph.ID, &ph.ProdutoID, &ph.URL, &ph.SortOrder, &ph.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar foto: %w", err)
	}
	return ph, nil
}

func DeletePhoto(photoID string) error {
	result, err := database.DB.Exec(
		`DELETE FROM produto_fotos WHERE id = $1`,
		photoID,
	)
	if err != nil {
		return fmt.Errorf("erro ao deletar foto: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("foto não encontrada")
	}
	return nil
}
