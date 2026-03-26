-- Revert offer_id back to store_id in prices
ALTER TABLE prices RENAME COLUMN offer_id TO store_id;

-- Revert FK constraint and index names
ALTER TABLE prices RENAME CONSTRAINT fk_prices_offer TO fk_prices_store;
ALTER INDEX idx_prices_offer_payment RENAME TO idx_prices_store_payment;

-- Rename offers table back to stores
ALTER TABLE offers RENAME TO stores;

-- Revert index name
ALTER INDEX idx_offers_product_id RENAME TO idx_stores_product_id;
