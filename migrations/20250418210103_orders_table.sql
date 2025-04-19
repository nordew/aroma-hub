-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50) NOT NULL,
    address TEXT NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    promo_code VARCHAR(50),
    contact_type VARCHAR(50) NOT NULL,
    amount_to_pay DECIMAL(12, 2) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW ()
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;

-- +goose StatementEnd
