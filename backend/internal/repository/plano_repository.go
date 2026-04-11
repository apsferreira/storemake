package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/apsferreira/storemake/backend/internal/domain"
)

type PlanoRepository struct {
	db *sql.DB
}

// NewPlanoRepository cria uma nova instância do repository de planos
func NewPlanoRepository(db *sql.DB) *PlanoRepository {
	return &PlanoRepository{db: db}
}

// GetBySlug busca um plano pelo slug
func (r *PlanoRepository) GetBySlug(ctx context.Context, slug string) (*domain.Plano, error) {
	query := `
		SELECT id, slug, name, price_cents, max_products, custom_domain, support_level, features, is_active, created_at, updated_at
		FROM planos
		WHERE slug = $1 AND is_active = true
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, slug)

	var plano domain.Plano
	var features sql.NullString

	err := row.Scan(
		&plano.ID, &plano.Slug, &plano.Name, &plano.PriceCents, &plano.MaxProducts,
		&plano.CustomDomain, &plano.SupportLevel, &features, &plano.IsActive,
		&plano.CreatedAt, &plano.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if features.Valid {
		plano.Features = json.RawMessage(features.String)
	}

	return &plano, nil
}

// GetByID busca um plano pelo ID
func (r *PlanoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Plano, error) {
	query := `
		SELECT id, slug, name, price_cents, max_products, custom_domain, support_level, features, is_active, created_at, updated_at
		FROM planos
		WHERE id = $1 AND is_active = true
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var plano domain.Plano
	var features sql.NullString

	err := row.Scan(
		&plano.ID, &plano.Slug, &plano.Name, &plano.PriceCents, &plano.MaxProducts,
		&plano.CustomDomain, &plano.SupportLevel, &features, &plano.IsActive,
		&plano.CreatedAt, &plano.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if features.Valid {
		plano.Features = json.RawMessage(features.String)
	}

	return &plano, nil
}

// ListActive retorna todos os planos ativos
func (r *PlanoRepository) ListActive(ctx context.Context) ([]domain.Plano, error) {
	query := `
		SELECT id, slug, name, price_cents, max_products, custom_domain, support_level, features, is_active, created_at, updated_at
		FROM planos
		WHERE is_active = true
		ORDER BY price_cents ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var planos []domain.Plano

	for rows.Next() {
		var plano domain.Plano
		var features sql.NullString

		err := rows.Scan(
			&plano.ID, &plano.Slug, &plano.Name, &plano.PriceCents, &plano.MaxProducts,
			&plano.CustomDomain, &plano.SupportLevel, &features, &plano.IsActive,
			&plano.CreatedAt, &plano.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if features.Valid {
			plano.Features = json.RawMessage(features.String)
		}

		planos = append(planos, plano)
	}

	return planos, rows.Err()
}

// CanCreateProduct verifica se uma loja pode criar um novo produto
// baseado no seu plano
func (r *PlanoRepository) CanCreateProduct(ctx context.Context, lojaID uuid.UUID) (bool, error) {
	query := `
		SELECT p.max_products - COALESCE((SELECT COUNT(*) FROM produtos WHERE loja_id = $1), 0) as remaining
		FROM lojas l
		JOIN planos p ON l.plano_id = p.id
		WHERE l.id = $1
		LIMIT 1
	`

	var remaining int
	err := r.db.QueryRowContext(ctx, query, lojaID).Scan(&remaining)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return remaining > 0, nil
}

// GetProductCount retorna o número de produtos de uma loja
func (r *PlanoRepository) GetProductCount(ctx context.Context, lojaID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM produtos WHERE loja_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, lojaID).Scan(&count)

	return count, err
}

// GetLojaPlano retorna o plano de uma loja
func (r *PlanoRepository) GetLojaPlano(ctx context.Context, lojaID uuid.UUID) (*domain.Plano, error) {
	query := `
		SELECT p.id, p.slug, p.name, p.price_cents, p.max_products, p.custom_domain,
		       p.support_level, p.features, p.is_active, p.created_at, p.updated_at
		FROM planos p
		JOIN lojas l ON l.plano_id = p.id
		WHERE l.id = $1
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, lojaID)

	var plano domain.Plano
	var features sql.NullString

	err := row.Scan(
		&plano.ID, &plano.Slug, &plano.Name, &plano.PriceCents, &plano.MaxProducts,
		&plano.CustomDomain, &plano.SupportLevel, &features, &plano.IsActive,
		&plano.CreatedAt, &plano.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if features.Valid {
		plano.Features = json.RawMessage(features.String)
	}

	return &plano, nil
}
