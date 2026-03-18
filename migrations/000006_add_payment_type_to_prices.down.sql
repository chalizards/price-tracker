DROP INDEX IF EXISTS idx_prices_product_payment;
ALTER TABLE prices DROP COLUMN IF EXISTS payment_type;
