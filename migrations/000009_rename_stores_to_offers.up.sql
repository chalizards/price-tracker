-- Rename stores table to offers
ALTER TABLE stores RENAME TO offers;

-- Rename indexes
ALTER INDEX idx_stores_product_id RENAME TO idx_offers_product_id;

-- Rename store_id column in prices to offer_id
ALTER TABLE prices RENAME COLUMN store_id TO offer_id;

-- Rename FK constraint and index
ALTER TABLE prices RENAME CONSTRAINT fk_prices_store TO fk_prices_offer;
ALTER INDEX idx_prices_store_payment RENAME TO idx_prices_offer_payment;
