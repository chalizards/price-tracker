ALTER TABLE prices ADD COLUMN payment_type VARCHAR(20) NOT NULL DEFAULT 'pix';

CREATE INDEX idx_prices_product_payment ON prices(product_id, payment_type, scraped_at DESC);
