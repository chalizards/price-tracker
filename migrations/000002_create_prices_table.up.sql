CREATE TABLE prices (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'BRL',
    scraped_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_prices_product_id ON prices(product_id);
CREATE INDEX idx_prices_scraped_at ON prices(scraped_at DESC);