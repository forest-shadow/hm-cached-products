package application

import (
	"context"
	"time"

	"github.com/forest-shadow/hm-cached-products/internal/product"
)

// ProductRepository provides product persistence operations to the application.
type ProductRepository interface {
	Create(context.Context, product.Input) (product.Product, error)
	Get(context.Context, int64) (product.Product, error)
	Update(context.Context, int64, int64, product.Input) (product.Product, error)
	Delete(context.Context, int64) error
}

// StatsReader is a read-side query; its result is not a Product aggregate.
type StatsReader interface {
	Stats(context.Context, int64, string, *time.Time) (product.Stats, error)
}
