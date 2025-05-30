-- +goose Up
-- +goose StatementBegin
-- First add primary key to products table if it doesn't exist
CREATE TABLE order_products (
    order_id UUID NOT NULL,
    product_id UUID NOT NULL,
    quantity INTEGER NOT NULL,
    volume INTEGER NOT NULL,
    PRIMARY KEY (order_id, product_id),
    FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products (id)
);

CREATE INDEX idx_order_products_product_id ON order_products (product_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS order_products;

-- +goose StatementEnd
