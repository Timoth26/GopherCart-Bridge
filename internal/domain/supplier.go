package domain

import (
	"context"
)

type SupplierClient interface {
	FetchProducts(ctx context.Context) ([]Product, error)
	SendOrder(ctx context.Context, order Order) error
}
