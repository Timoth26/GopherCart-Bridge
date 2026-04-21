package domain

import (
	"context"
	"time"
)

type OrderStatus string

const (
	OrderStatusPending OrderStatus = "pending"
	OrderStatusSent    OrderStatus = "sent"
)

type Order struct {
	ID         int64       `json:"id"          db:"id"`
	TotalPrice float64     `json:"total_price" db:"total_price"`
	Status     OrderStatus `json:"status"      db:"status"`
	CreatedAt  time.Time   `json:"created_at"  db:"created_at"`
	Items      []OrderItem `json:"items"       db:"-"`
}

type OrderItem struct {
	ID        int64   `json:"id"         db:"id"`
	OrderID   int64   `json:"order_id"   db:"order_id"`
	ProductID int64   `json:"product_id" db:"product_id"`
	Quantity  int     `json:"quantity"   db:"quantity"`
	LineTotal float64 `json:"line_total" db:"line_total"`
	Product   Product `json:"product"    db:"-"`
}

type OrderRepository interface {
	GetByID(ctx context.Context, id int64) (*Order, error)
	GetAll(ctx context.Context) ([]Order, error)
	Create(ctx context.Context, o *Order) error
	UpdateStatus(ctx context.Context, id int64, status OrderStatus) error
	Delete(ctx context.Context, id int64) error
}
