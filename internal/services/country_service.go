package internal

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"

	db "github.com/franzego/stage02/db/sqlc"
	"github.com/franzego/stage02/models"
)

type CountryService struct {
	q           *db.Queries
	externalapi *ExternalApi
}

func NewCountryService(queries *db.Queries) *CountryService {
	return &CountryService{
		q:           queries,
		externalapi: NewExternalService(),
	}
}

// function to get totalcount
func (c *CountryService) GetTotalCount() (int64, error) {
	ctx := context.Background()
	count, err := c.q.GetTotalCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("could not get the total countries: %w", err)
	}
	return count, nil
}

// function to get refreshtime
func (c *CountryService) GetRefreshTime() (sql.NullTime, error) {
	ctx := context.Background()
	t, err := c.q.GetLatestRefreshTime(ctx)
	if err != nil {
		return sql.NullTime{}, fmt.Errorf("there was a problem getting refresh time: %w", err)
	}
	return t, nil
}

// function to get all countries
func (c *CountryService) GetAllCountries() ([]db.Country, error) {
	ctx := context.Background()
	countries, err := c.q.GetAllCountries(ctx)
	if err != nil {
		return nil, fmt.Errorf("external data source unavailable: %w", err)
	}
	return countries, nil
}

// function to get countries by name
func (c *CountryService) GetCountryByName(name string) (db.Country, error) {
	ctx := context.Background()
	country, err := c.q.GetCountryByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.Country{}, err
		}
	}
	return country, nil

}

// function to delete countries by name
func (c *CountryService) DeleteCountryByName(name string) error {
	// check if it exists in db
	ctx := context.Background()
	country, err := c.q.GetCountryByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return err
		}
		return err
	}
	// if there is no error, and it exists, we then delete
	if err = c.q.DeleteCountryByName(ctx, country.Name); err != nil {
		return err
	}
	return nil
}

// function to refresh countries
func (c *CountryService) RefreshCountries() error {
	fmt.Println("Starting country refresh...")
	country, err := c.externalapi.FetchAllCountries()
	if err != nil {
		fmt.Printf("Error fetching countries: %v\n", err)
		return fmt.Errorf("external data source unavailable: %w", err)
	}
	fmt.Printf("Successfully fetched %d countries\n", len(country))

	fmt.Println("Fetching exchange rates...")
	rates, err := c.externalapi.FetchExchangeRate()
	if err != nil {
		fmt.Printf("Error fetching exchange rates: %v\n", err)
		return fmt.Errorf("external rates source unavailable: %w", err)
	}
	fmt.Printf("Successfully fetched exchange rates for %d currencies\n", len(rates))

	ctx := context.Background()
	for _, count := range country {
		processed := c.processCountry(count, rates)
		if err := c.upsertCountry(ctx, processed); err != nil {
			fmt.Printf("Failed to upsert country %s: %v\n", processed.Name, err)
			continue
		}
	}
	imageService := NewImageService(c.q)
	if err := imageService.GenerateSummaryImage(ctx); err != nil {
		// Log error but don't fail the whole refresh
		fmt.Printf("Failed to generate image: %v\n", err)
	}

	return nil
}

// function that processes data
// processCountry extracts currency, calculates GDP, and prepares data
func (c *CountryService) processCountry(country models.CountryData, exchangeRates map[string]float64) models.ProcessedCountry {
	processed := models.ProcessedCountry{
		Name:       country.Name,
		Capital:    country.Capital,
		Region:     country.Region,
		Population: country.Population,
		FlagURL:    country.Flag,
	}

	// Extract first currency code
	if len(country.Currencies) > 0 && country.Currencies[0].Code != "" {
		currencyCode := country.Currencies[0].Code
		processed.CurrencyCode = &currencyCode

		// Match with exchange rate
		if rate, exists := exchangeRates[currencyCode]; exists && rate > 0 {
			processed.ExchangeRate = &rate

			// Calculate estimated GDP
			gdp := c.calculateGDP(country.Population, rate)
			processed.EstimatedGDP = &gdp
		} else {
			// Currency not found in exchange rates
			processed.ExchangeRate = nil
			processed.EstimatedGDP = nil
		}
	} else {
		// No currency for this country
		processed.CurrencyCode = nil
		processed.ExchangeRate = nil
		gdp := 0.0
		processed.EstimatedGDP = &gdp
	}

	return processed
}

// function to calculate gdp
func (c *CountryService) calculateGDP(population int64, rate float64) float64 {
	multiplier := rand.Float64()*1000 + 1000

	gdp := (float64(population) * multiplier) / rate
	return gdp
}

// function to insert into db
func (c *CountryService) upsertCountry(ctx context.Context, country models.ProcessedCountry) error {
	// Convert nullable fields to sql.Null types
	var capital, region, currencyCode, flagURL sql.NullString
	var exchangeRate, estimatedGDP sql.NullString

	if country.Capital != "" {
		capital = sql.NullString{String: country.Capital, Valid: true}
	}

	if country.Region != "" {
		region = sql.NullString{String: country.Region, Valid: true}
	}

	if country.CurrencyCode != nil {
		currencyCode = sql.NullString{String: *country.CurrencyCode, Valid: true}
	}

	if country.FlagURL != "" {
		flagURL = sql.NullString{String: country.FlagURL, Valid: true}
	}

	if country.ExchangeRate != nil {
		exchangeRate = sql.NullString{String: strconv.FormatFloat(*country.ExchangeRate, 'f', 6, 64), Valid: true}
	}

	if country.EstimatedGDP != nil {
		estimatedGDP = sql.NullString{String: strconv.FormatFloat(*country.EstimatedGDP, 'f', 2, 64), Valid: true}
	}

	return c.q.UpsertCountry(ctx, db.UpsertCountryParams{
		Name:         country.Name,
		Capital:      capital,
		Region:       region,
		Population:   country.Population,
		CurrencyCode: currencyCode,
		ExchangeRate: exchangeRate,
		EstimatedGdp: estimatedGDP,
		FlagUrl:      flagURL,
	})
}
