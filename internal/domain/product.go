package domain

import "context"

type Product struct {
	ID          int64   `json:"id"          db:"id"`
	Name        string  `json:"name"        db:"name"`
	Description string  `json:"description" db:"description"`
	Price       float64 `json:"price"       db:"price"`
	Stock       int     `json:"stock"       db:"stock"`
	ProviderID  int     `json:"provider_id" db:"provider_id"`
}

type ProductRepository interface {
	GetByID(ctx context.Context, id int64) (*Product, error)
	GetAll(ctx context.Context) ([]Product, error)
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id int64) error
}
