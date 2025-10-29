CREATE TABLE IF NOT EXISTS countries (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL UNIQUE,
    capital VARCHAR(255) NULL,
    region VARCHAR(100) NULL,
    population BIGINT NOT NULL,
    currency_code VARCHAR(10) NULL,
    exchange_rate DECIMAL(20, 6) NULL,
    estimated_gdp DECIMAL(30, 2) NULL,
    flag_url TEXT NULL,
    last_refreshed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_region (region),
    INDEX idx_currency (currency_code)
);