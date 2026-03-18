ALTER TABLE products ADD COLUMN slug VARCHAR(255) NOT NULL DEFAULT '';
CREATE UNIQUE INDEX idx_products_slug ON products(slug);
