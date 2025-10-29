-- name: UpsertCountry :exec
INSERT INTO countries (
    name, capital, region, population, 
    currency_code, exchange_rate, estimated_gdp, flag_url, last_refreshed_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
ON DUPLICATE KEY UPDATE
    capital = VALUES(capital),
    region = VALUES(region),
    population = VALUES(population),
    currency_code = VALUES(currency_code),
    exchange_rate = VALUES(exchange_rate),
    estimated_gdp = VALUES(estimated_gdp),
    flag_url = VALUES(flag_url),
    last_refreshed_at = NOW();

-- name: GetAllCountries :many
SELECT * FROM countries
ORDER BY id;

-- name: GetCountryByName :one
SELECT * FROM countries
WHERE LOWER(name) = LOWER(?);

-- name: DeleteCountryByName :exec
DELETE FROM countries WHERE LOWER(name) = LOWER(?);

-- name: GetTotalCount :one
SELECT COUNT(*) as total FROM countries;

-- name: GetLatestRefreshTime :one
SELECT last_refreshed_at as last_refresh FROM countries
ORDER BY last_refreshed_at DESC
LIMIT 1;

-- name: GetTopCountriesByGDP :many
SELECT * FROM countries
WHERE estimated_gdp IS NOT NULL
ORDER BY estimated_gdp DESC
LIMIT ?;
