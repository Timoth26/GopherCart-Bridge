package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	"github.com/Timoth26/GopherCart-Bridge/internal/domain"
)

const pgDuplicateKey = "23505"

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	var p domain.Product

	err := r.db.GetContext(ctx, &p, `
		SELECT id, name, description, price, stock, provider_id
		FROM products
		WHERE id = $1
	`, id)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get product by id: %w", err)
	}

	return &p, nil
}

func (r *ProductRepository) GetAll(ctx context.Context) ([]domain.Product, error) {
	var products []domain.Product

	err := r.db.SelectContext(ctx, &products, `
		SELECT id, name, description, price, stock, provider_id
		FROM products
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("get all products: %w", err)
	}

	return products, nil
}

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO products (name, description, price, stock, provider_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, p.Name, p.Description, p.Price, p.Stock, p.ProviderID).Scan(&p.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgDuplicateKey {
			return domain.ErrProductAlreadyExists
		}
		return fmt.Errorf("create product: %w", err)
	}

	return nil
}

func (r *ProductRepository) Update(ctx context.Context, p *domain.Product) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock = $4, provider_id = $5, updated_at = NOW()
		WHERE id = $6
	`, p.Name, p.Description, p.Price, p.Stock, p.ProviderID, p.ID)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update product rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM products
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete product rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}
