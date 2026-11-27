package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/forest-shadow/hm-cached-products/internal/product"
)

var (
	ErrNotFound      = errors.New("product not found")
	ErrConflict      = errors.New("product version conflict")
	ErrInvalidWindow = errors.New("invalid statistics window")
)

var statsWindowDurations = map[string]time.Duration{
	"1h":  time.Hour,
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
}

type Service struct {
	products ProductRepository
	stats    StatsReader
	now      func() time.Time
}

func NewService(products ProductRepository, stats StatsReader) *Service {
	return &Service{products: products, stats: stats, now: time.Now}
}

func (s *Service) Create(ctx context.Context, input product.Input) (product.Product, error) {
	const op = "application.Service.Create"

	clean, err := product.Validate(input)
	if err != nil {
		return product.Product{}, fmt.Errorf("%s: validate product: %w", op, err)
	}

	created, err := s.products.Create(ctx, clean)
	if err != nil {
		return product.Product{}, fmt.Errorf("%s: persist product: %w", op, err)
	}

	return created, nil
}

func (s *Service) Get(ctx context.Context, id int64) (product.Product, error) {
	const op = "application.Service.Get"

	if id <= 0 {
		return product.Product{}, fmt.Errorf("%s: validate product ID: %w", op, product.ErrInvalid)
	}

	found, err := s.products.Get(ctx, id)
	if err != nil {
		return product.Product{}, fmt.Errorf("%s: load product: %w", op, err)
	}

	return found, nil
}

func (s *Service) Update(ctx context.Context, command UpdateProductCommand) (product.Product, error) {
	const op = "application.Service.Update"

	if err := command.validate(); err != nil {
		return product.Product{}, fmt.Errorf("%s: validate command: %w", op, err)
	}

	updated, err := s.products.Update(
		ctx,
		command.ProductID(),
		command.ExpectedVersion(),
		command.Input(),
	)
	if err != nil {
		return product.Product{}, fmt.Errorf("%s: persist product: %w", op, err)
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	const op = "application.Service.Delete"

	if id <= 0 {
		return fmt.Errorf("%s: validate product ID: %w", op, product.ErrInvalid)
	}

	if err := s.products.Delete(ctx, id); err != nil {
		return fmt.Errorf("%s: delete product: %w", op, err)
	}

	return nil
}

func (s *Service) Stats(ctx context.Context, id int64, window string) (product.Stats, error) {
	const op = "application.Service.Stats"

	if id <= 0 {
		return product.Stats{}, fmt.Errorf("%s: validate product ID: %w", op, product.ErrInvalid)
	}

	if window == "" {
		window = "24h"
	}

	var cutoff *time.Time
	if duration, ok := statsWindowDurations[window]; ok {
		t := s.now().Add(-duration)
		cutoff = &t
	} else if window != "all" {
		return product.Stats{}, fmt.Errorf("%s: %w", op, ErrInvalidWindow)
	}

	stats, err := s.stats.Stats(ctx, id, window, cutoff)
	if err != nil {
		return product.Stats{}, fmt.Errorf("%s: load product statistics: %w", op, err)
	}
	stats.Window = window

	return stats, nil
}
