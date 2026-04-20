package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/Timoth26/GopherCart-Bridge/internal/domain"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	var o domain.Order

	err := r.db.GetContext(ctx, &o, `
		SELECT id, total_price, status, created_at
		FROM orders
		WHERE id = $1
	`, id)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get order by id: %w", err)
	}

	items, err := r.fetchItems(ctx, id)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *OrderRepository) GetAll(ctx context.Context) ([]domain.Order, error) {
	var orders []domain.Order

	err := r.db.SelectContext(ctx, &orders, `
		SELECT id, total_price, status, created_at
		FROM orders
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("get all orders: %w", err)
	}

	if len(orders) == 0 {
		return orders, nil
	}

	ids := make([]int64, len(orders))
	for i, o := range orders {
		ids[i] = o.ID
	}

	itemsByOrder, err := r.fetchItemsForOrders(ctx, ids)
	if err != nil {
		return nil, err
	}

	for i := range orders {
		orders[i].Items = itemsByOrder[orders[i].ID]
	}

	return orders, nil
}

func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) (err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("create order begin tx: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			err = errors.Join(err, fmt.Errorf("rollback: %w", rbErr))
		}
	}()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO orders (total_price, status)
		VALUES ($1, $2)
		RETURNING id, created_at
	`, o.TotalPrice, o.Status).Scan(&o.ID, &o.CreatedAt)
	if err != nil {
		return fmt.Errorf("create order insert: %w", err)
	}

	for i := range o.Items {
		o.Items[i].OrderID = o.ID
		err = tx.QueryRowContext(ctx, `
			INSERT INTO order_items (order_id, product_id, quantity, line_total)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, o.Items[i].OrderID, o.Items[i].ProductID, o.Items[i].Quantity, o.Items[i].LineTotal).
			Scan(&o.Items[i].ID)
		if err != nil {
			return fmt.Errorf("create order item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("create order commit: %w", err)
	}

	return nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE orders
		SET status = $1
		WHERE id = $2
	`, status, id)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update order status rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func (r *OrderRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM orders
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete order: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete order rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func (r *OrderRepository) fetchItems(ctx context.Context, orderID int64) ([]domain.OrderItem, error) {
	byOrder, err := r.fetchItemsForOrders(ctx, []int64{orderID})
	if err != nil {
		return nil, err
	}
	return byOrder[orderID], nil
}

func (r *OrderRepository) fetchItemsForOrders(ctx context.Context, orderIDs []int64) (map[int64][]domain.OrderItem, error) {
	type row struct {
		ID        int64   `db:"id"`
		OrderID   int64   `db:"order_id"`
		ProductID int64   `db:"product_id"`
		Quantity  int     `db:"quantity"`
		LineTotal float64 `db:"line_total"`

		ProdName        string  `db:"prod_name"`
		ProdDescription string  `db:"prod_description"`
		ProdPrice       float64 `db:"prod_price"`
		ProdStock       int     `db:"prod_stock"`
		ProdProviderID  int     `db:"prod_provider_id"`
	}

	query, args, err := sqlx.In(`
		SELECT
			oi.id,
			oi.order_id,
			oi.product_id,
			oi.quantity,
			oi.line_total,
			p.name        AS prod_name,
			p.description AS prod_description,
			p.price       AS prod_price,
			p.stock       AS prod_stock,
			p.provider_id AS prod_provider_id
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id IN (?)
		ORDER BY oi.order_id, oi.id
	`, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("build fetch items query: %w", err)
	}
	query = r.db.Rebind(query)

	var rows []row
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("fetch order items: %w", err)
	}

	result := make(map[int64][]domain.OrderItem, len(orderIDs))
	for _, r := range rows {
		result[r.OrderID] = append(result[r.OrderID], domain.OrderItem{
			ID:        r.ID,
			OrderID:   r.OrderID,
			ProductID: r.ProductID,
			Quantity:  r.Quantity,
			LineTotal: r.LineTotal,
			Product: domain.Product{
				ID:          r.ProductID,
				Name:        r.ProdName,
				Description: r.ProdDescription,
				Price:       r.ProdPrice,
				Stock:       r.ProdStock,
				ProviderID:  r.ProdProviderID,
			},
		})
	}

	return result, nil
}
