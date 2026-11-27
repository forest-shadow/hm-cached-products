package productsrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/forest-shadow/hm-cached-products/internal/application"
	"github.com/forest-shadow/hm-cached-products/internal/product"
)

type Repository struct {
	db               *sqlx.DB
	operationTimeout time.Duration
}

// New creates a product repository using the provided database pool.
func New(db *sqlx.DB, operationTimeout time.Duration) *Repository {
	return &Repository{db: db, operationTimeout: operationTimeout}
}

func (r *Repository) Create(ctx context.Context, input product.Input) (product.Product, error) {
	const op = "postgres.products.Repository.Create"

	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()

	var row productRow
	err := r.db.GetContext(ctx, &row, `
		INSERT INTO products (name, price)
		VALUES ($1, $2)
		RETURNING id, name, price, version, updated_at
	`, input.Name, input.Price)
	if err != nil {
		return product.Product{}, fmt.Errorf("%s: insert product: %w", op, err)
	}

	return row.toProduct(), err
}

func (r *Repository) Get(ctx context.Context, id int64) (product.Product, error) {
	const op = "postgres.products.Repository.Get"

	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()

	var row productRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, name, price, version, updated_at
		FROM products WHERE id = $1
	`, id)

	if errors.Is(err, sql.ErrNoRows) {
		return product.Product{}, fmt.Errorf("%s: %w", op, application.ErrNotFound)
	}
	if err != nil {
		return product.Product{}, fmt.Errorf("%s: select product: %w", op, err)
	}

	return row.toProduct(), nil
}

func (r *Repository) Update(ctx context.Context, id, expectedVersion int64, input product.Input) (product.Product, error) {
	const op = "postgres.products.Repository.Update"

	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return product.Product{}, fmt.Errorf("%s: begin transaction: %w", op, err)
	}
	defer tx.Rollback()

	var row productRow
	err = tx.GetContext(ctx, &row, `
		SELECT id, name, price, version, updated_at
		FROM products WHERE id = $1 FOR UPDATE
	`, id)

	if errors.Is(err, sql.ErrNoRows) {
		return product.Product{}, fmt.Errorf("%s: %w", op, application.ErrNotFound)
	}
	if err != nil {
		return product.Product{}, fmt.Errorf("%s: lock product: %w", op, err)
	}
	// Check the version before comparing fields: even an unchanged request
	// must not succeed against a stale version of the product.
	if row.Version != expectedVersion {
		return product.Product{}, fmt.Errorf("%s: %w", op, application.ErrConflict)
	}

	if row.Name != input.Name || row.Price != input.Price {
		err = tx.GetContext(ctx, &row, `
			UPDATE products
			SET name = $3, price = $4, version = version + 1, updated_at = now()
			WHERE id = $1 AND version = $2
			RETURNING id, name, price, version, updated_at
		`, id, expectedVersion, input.Name, input.Price)
		if err != nil {
			return product.Product{}, fmt.Errorf("%s: update product: %w", op, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return product.Product{}, fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return row.toProduct(), nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const op = "postgres.products.Repository.Delete"

	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: begin transaction: %w", op, err)
	}
	defer tx.Rollback()

	var deletedVersion int64
	err = tx.GetContext(ctx, &deletedVersion, `
		DELETE FROM products WHERE id = $1
		RETURNING version
	`, id)

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", op, application.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: delete product: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return nil
}

func (r *Repository) Stats(ctx context.Context, id int64, _ string, cutoff *time.Time) (product.Stats, error) {
	const op = "postgres.products.Repository.Stats"

	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()

	var row statsRow
	err := r.db.GetContext(ctx, &row, `
		SELECT 
			p.id AS product_id,
			COUNT(v.product_id) AS view_count,
			COUNT(DISTINCT (v.viewed_at AT TIME ZONE 'UTC')::date) AS unique_view_days,
			COUNT(v.product_id) FILTER (WHERE v.viewed_at >= p.updated_at) AS views_since_update,
			MIN(v.viewed_at) AS first_viewed_at,
			MAX(v.viewed_at) AS last_viewed_at
		FROM products p
		LEFT JOIN product_views v ON v.product_id = p.id
			AND ($2::timestamptz IS NULL OR v.viewed_at >= $2::timestamptz)
		WHERE p.id = $1 GROUP BY p.id, p.updated_at
	`, id, cutoff)

	if errors.Is(err, sql.ErrNoRows) {
		return product.Stats{}, fmt.Errorf("%s: %w", op, application.ErrNotFound)
	}
	if err != nil {
		return product.Stats{}, fmt.Errorf("%s: select product statistics: %w", op, err)
	}

	return row.toStats(), nil
}
