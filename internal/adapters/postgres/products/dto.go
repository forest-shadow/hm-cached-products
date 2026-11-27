package productsrepo

import (
	"time"

	"github.com/forest-shadow/hm-cached-products/internal/product"
)

type productRow struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Price     int64     `db:"price"`
	Version   int64     `db:"version"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r productRow) toProduct() product.Product {
	return product.Product{
		ID: r.ID,
		Name: r.Name,
		Price: r.Price,
		Version: r.Version,
		UpdatedAt: r.UpdatedAt,
	}
}

type statsRow struct {
	ProductID        int64      `db:"product_id"`
	ViewCount        int64      `db:"view_count"`
	UniqueViewDays   int64      `db:"unique_view_days"`
	ViewsSinceUpdate int64      `db:"views_since_update"`
	FirstViewedAt    *time.Time `db:"first_viewed_at"`
	LastViewedAt     *time.Time `db:"last_viewed_at"`
}

func (r statsRow) toStats() product.Stats {
	return product.Stats{
		ProductID: r.ProductID,
		ViewCount: r.ViewCount,
		UniqueViewDays: r.UniqueViewDays,
		ViewsSinceUpdate: r.ViewsSinceUpdate,
		FirstViewedAt: r.FirstViewedAt, 
		LastViewedAt: r.LastViewedAt,
	}
}
