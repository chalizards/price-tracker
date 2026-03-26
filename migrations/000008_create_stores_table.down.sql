-- Re-add url and store columns to products
ALTER TABLE products ADD COLUMN url TEXT;
ALTER TABLE products ADD COLUMN store VARCHAR(255);

-- Restore product data from stores
UPDATE products p
SET url = s.url, store = s.name
FROM stores s
WHERE s.product_id = p.id;

-- Re-add product_id to prices
ALTER TABLE prices ADD COLUMN product_id INT;

-- Restore product_id from store mapping
UPDATE prices pr
SET product_id = s.product_id
FROM stores s
WHERE s.id = pr.store_id;

ALTER TABLE prices ALTER COLUMN product_id SET NOT NULL;
ALTER TABLE prices ADD CONSTRAINT fk_prices_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

-- Drop store_id from prices
ALTER TABLE prices DROP COLUMN store_id;
DROP INDEX IF EXISTS idx_prices_store_id;

-- Drop stores table
DROP TABLE IF EXISTS stores;
