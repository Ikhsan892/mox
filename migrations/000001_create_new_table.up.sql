
-- Create a new schema
CREATE SCHEMA IF NOT EXISTS public;

-- Create a table within the new schema
CREATE TABLE orders (
    id VARCHAR(26) PRIMARY KEY,
    customer_name VARCHAR(255) NOT NULL,
    total_amount NUMERIC NOT NULL,
    status VARCHAR(255) NOT NULL,
    address TEXT,
    created_at timestamp NOT NULL,
    created bigint not null,
    updated BIGINT NOT NULL,
    deleted_at BIGINT
);


CREATE TABLE order_items (
    id VARCHAR(26) PRIMARY KEY,
    product_id VARCHAR(255) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    price NUMERIC,
    created_at timestamp NOT NULL,
    created bigint not null,
    updated BIGINT NOT NULL,
    deleted_at BIGINT
);
