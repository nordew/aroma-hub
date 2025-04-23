-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS admins (
    id UUID NOT NULL,
    vendor_id VARCHAR(255) NOT NULL,
    vendor_type VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    UNIQUE (vendor_id, vendor_type)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS admins;

-- +goose StatementEnd
