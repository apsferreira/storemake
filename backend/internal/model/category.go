package model

import "time"

type Category struct {
	ID        string    `json:"id"`
	LojaID    string    `json:"loja_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateCategoryRequest struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	SortOrder int    `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name      *string `json:"name,omitempty"`
	Slug      *string `json:"slug,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
}
