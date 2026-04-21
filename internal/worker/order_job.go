package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Timoth26/GopherCart-Bridge/internal/domain"
)

type orderJob struct {
	supplier domain.SupplierClient
	orders   orderService
	log      *slog.Logger
}

func newOrderJob(supplier domain.SupplierClient, orders orderService, log *slog.Logger) *orderJob {
	return &orderJob{
		supplier: supplier,
		orders:   orders,
		log:      log,
	}
}

func (j *orderJob) run(ctx context.Context) error {
	orders, err := j.orders.GetPending(ctx)
	if err != nil {
		return fmt.Errorf("get pending orders: %w", err)
	}

	var failed int
	for _, order := range orders {
		if err := j.supplier.SendOrder(ctx, order); err != nil {
			j.log.ErrorContext(ctx, "send order failed", "order_id", order.ID, "err", err)
			failed++
			continue
		}
		if err := j.orders.UpdateStatus(ctx, order.ID, "sent"); err != nil {
			j.log.ErrorContext(ctx, "update order status failed", "order_id", order.ID, "err", err)
			failed++
			continue
		}
	}

	j.log.InfoContext(ctx, "order job completed",
		"total", len(orders),
		"ok", len(orders)-failed,
		"failed", failed,
	)
	return nil
}
