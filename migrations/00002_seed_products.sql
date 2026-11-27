-- +goose Up
INSERT INTO products (id, name, price, version, updated_at) VALUES
    (1, 'No views product', 1_990, 1, now() - INTERVAL '7 days'),
    (2, 'Recently viewed product', 2_490, 1, now() - INTERVAL '1 day'),
    (3, 'Popular product', 3_990, 1, now() - INTERVAL '2 hours'),
    (4, 'Historical views product', 4_990, 3, now() - INTERVAL '30 minutes'),
    (5, 'Updated product', 5_990, 7, now() - INTERVAL '5 minutes'),
    (6, 'Free product', 0, 1, now() - INTERVAL '3 days'),
    (7, 'High price product', 9_223_372_036_854_775_807, 1, now() - INTERVAL '1 hour');

-- One fresh view: a non-empty, low-volume result.
INSERT INTO product_views (product_id, viewed_at)
VALUES (2, now() - INTERVAL '10 minutes');

-- Many fresh views: suitable for cache, aggregation, and concurrent-load checks.
INSERT INTO product_views (product_id, viewed_at)
SELECT 3, now() - (n * INTERVAL '1 minute')
FROM generate_series(0, 59) AS n;

-- Views in several time windows, plus one expired view.
INSERT INTO product_views (product_id, viewed_at) VALUES
    (4, now() - INTERVAL '5 minutes'),
    (4, now() - INTERVAL '2 hours'),
    (4, now() - INTERVAL '2 days'),
    (4, now() - INTERVAL '14 days'),
    (4, now() - INTERVAL '90 days');

-- A recently changed product with views before and after its last update.
INSERT INTO product_views (product_id, viewed_at) VALUES
    (5, now() - INTERVAL '10 minutes'),
    (5, now() - INTERVAL '1 minute');

SELECT setval(pg_get_serial_sequence('products', 'id'), 7, true);

-- +goose Down
DELETE FROM product_views
WHERE product_id IN (
    SELECT id
    FROM products
    WHERE id BETWEEN 1 AND 7
      AND name IN (
          'No views product',
          'Recently viewed product',
          'Popular product',
          'Historical views product',
          'Updated product',
          'Free product',
          'High price product'
      )
);

DELETE FROM products
WHERE id BETWEEN 1 AND 7
  AND name IN (
      'No views product',
      'Recently viewed product',
      'Popular product',
      'Historical views product',
      'Updated product',
      'Free product',
      'High price product'
  );
