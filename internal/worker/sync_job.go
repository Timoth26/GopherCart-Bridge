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
	cache    domain.ProductCache
	log      *slog.Logger
}

func newSyncJob(
	supplier domain.SupplierClient,
	products productService,
	cache domain.ProductCache,
	log *slog.Logger,
) *syncJob {
	return &syncJob{
		supplier: supplier,
		products: products,
		cache:    cache,
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
		if err := j.cache.Delete(ctx, products[i].ID); err != nil {
			j.log.WarnContext(ctx, "cache invalidation failed", "product_id", products[i].ID, "err", err)
		}
	}

	j.log.InfoContext(ctx, "sync job completed",
		"total", len(products),
		"ok", len(products)-failed,
		"failed", failed,
	)
	return nil
}
