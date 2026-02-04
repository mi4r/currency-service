-- Currency rates table
CREATE TABLE IF NOT EXISTS currency_rates (
    id              BIGSERIAL PRIMARY KEY,
    rate_date       DATE NOT NULL,
    base_currency   VARCHAR(3) NOT NULL DEFAULT 'RUB',
    target_currency VARCHAR(10) NOT NULL,
    rate            DECIMAL(38, 18) NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Composite unique constraint
    CONSTRAINT uq_rate_date_currencies
        UNIQUE (rate_date, base_currency, target_currency)
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_currency_rates_date
    ON currency_rates(rate_date);

CREATE INDEX IF NOT EXISTS idx_currency_rates_target
    ON currency_rates(target_currency);

CREATE INDEX IF NOT EXISTS idx_currency_rates_date_target
    ON currency_rates(rate_date, target_currency);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for auto-updating updated_at
DROP TRIGGER IF EXISTS update_currency_rates_updated_at ON currency_rates;
CREATE TRIGGER update_currency_rates_updated_at
    BEFORE UPDATE ON currency_rates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
