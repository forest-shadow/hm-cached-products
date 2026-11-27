package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/forest-shadow/hm-cached-products/internal/product"
)

type productResponseBody struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Price     int64     `json:"price"`
	Version   int64     `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

func productResponse(p product.Product) productResponseBody {
	return productResponseBody{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		Version:   p.Version,
		UpdatedAt: p.UpdatedAt,
	}
}

type statsResponseBody struct {
	ProductID        int64      `json:"product_id"`
	Window           string     `json:"window"`
	ViewCount        int64      `json:"view_count"`
	UniqueViewDays   int64      `json:"unique_view_days"`
	ViewsSinceUpdate int64      `json:"views_since_update"`
	FirstViewedAt    *time.Time `json:"first_viewed_at"`
	LastViewedAt     *time.Time `json:"last_viewed_at"`
}

func statsResponse(s product.Stats) statsResponseBody {
	return statsResponseBody{
		ProductID:        s.ProductID,
		Window:           s.Window,
		ViewCount:        s.ViewCount,
		UniqueViewDays:   s.UniqueViewDays,
		ViewsSinceUpdate: s.ViewsSinceUpdate,
		FirstViewedAt:    s.FirstViewedAt,
		LastViewedAt:     s.LastViewedAt,
	}
}

type healthResponse struct {
	Status string `json:"status"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
