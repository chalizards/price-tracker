-- Create stores table
CREATE TABLE stores (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stores_product_id ON stores(product_id);

-- Migrate existing product data into stores
INSERT INTO stores (product_id, name, url)
SELECT id, store, url FROM products WHERE store IS NOT NULL AND url IS NOT NULL;

-- Add store_id column to prices
ALTER TABLE prices ADD COLUMN store_id INT;

-- Populate store_id from existing product_id mapping
UPDATE prices p
SET store_id = s.id
FROM stores s
WHERE s.product_id = p.product_id;

-- Make store_id NOT NULL and add FK
ALTER TABLE prices ALTER COLUMN store_id SET NOT NULL;
ALTER TABLE prices ADD CONSTRAINT fk_prices_store FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE;

CREATE INDEX idx_prices_store_id ON prices(store_id);

-- Drop old product_id column from prices
ALTER TABLE prices DROP COLUMN product_id;

-- Remove store and url columns from products
ALTER TABLE products DROP COLUMN IF EXISTS store;
ALTER TABLE products DROP COLUMN IF EXISTS url;
