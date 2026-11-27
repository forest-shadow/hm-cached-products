package product

import "time"

// Stats contains product view aggregates for a requested time window.
type Stats struct {
	ProductID        int64
	Window           string
	ViewCount        int64
	UniqueViewDays   int64
	ViewsSinceUpdate int64
	FirstViewedAt    *time.Time
	LastViewedAt     *time.Time
}
