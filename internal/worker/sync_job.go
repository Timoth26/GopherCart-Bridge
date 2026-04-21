package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Timoth26/GopherCart-Bridge/internal/domain"
)

type syncJob struct {
	supplier domain.SupplierClient
	products productService
	log      *slog.Logger
}

func newSyncJob(supplier domain.SupplierClient, products productService, log *slog.Logger) *syncJob {
	return &syncJob{
		supplier: supplier,
		products: products,
		log:      log,
	}
}

func (j *syncJob) run(ctx context.Context) error {
	products, err := j.supplier.FetchProducts(ctx)
	if err != nil {
		return fmt.Errorf("fetch products from supplier: %w", err)
	}

	var failed int
	for i := range products {
		if err := j.products.Upsert(ctx, &products[i]); err != nil {
			j.log.ErrorContext(ctx, "upsert product failed", "product_id", products[i].ID, "err", err)
			failed++
			continue
		}
	}

	j.log.InfoContext(ctx, "sync job completed",
		"total", len(products),
		"ok", len(products)-failed,
		"failed", failed,
	)
	return nil
}
