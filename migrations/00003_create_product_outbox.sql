-- +goose Up
CREATE TABLE product_outbox (
    event_id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_type       TEXT NOT NULL CHECK (event_type IN ('product.created', 'product.updated', 'product.deleted')),
    product_id       BIGINT NOT NULL CHECK (product_id > 0),
    product_version  BIGINT NOT NULL CHECK (product_version > 0),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    available_at     TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    attempts         INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    claim_token      UUID,
    claimed_until    TIMESTAMPTZ,
    processed_at     TIMESTAMPTZ,
    last_error       VARCHAR(512),
    CONSTRAINT product_outbox_claim_pair CHECK ((claim_token IS NULL) = (claimed_until IS NULL))
);

-- product_id intentionally has no FK: a delete event must survive removal of its product.
CREATE INDEX product_outbox_available_idx
    ON product_outbox (available_at, event_id)
    WHERE processed_at IS NULL;

CREATE INDEX product_outbox_processed_idx
    ON product_outbox (processed_at)
    WHERE processed_at IS NOT NULL;

-- This is called separately from request handling and event delivery. Passing
-- a cutoff keeps the retention period an operational choice.
-- +goose StatementBegin
CREATE FUNCTION purge_processed_product_outbox(before_time TIMESTAMPTZ)
RETURNS BIGINT LANGUAGE sql AS $$
    WITH deleted AS (
        DELETE FROM product_outbox
        WHERE processed_at < before_time
        RETURNING 1
    )
    SELECT count(*) FROM deleted
$$;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION purge_processed_product_outbox(TIMESTAMPTZ);
DROP TABLE product_outbox;
