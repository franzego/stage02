package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/franzego/stage02/models"
)

type ExternalApi struct {
	httpclient *http.Client
}

func NewExternalService() *ExternalApi {
	return &ExternalApi{
		httpclient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (e *ExternalApi) FetchAllCountries() ([]models.CountryData, error) {
	url := "https://restcountries.com/v2/all?fields=name,capital,region,population,flag,currencies"
	resp, err := e.httpclient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch countries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("restcountries API returned status %d", resp.StatusCode)
	}
	var countries []models.CountryData
	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		return nil, fmt.Errorf("failed to parse countries JSON: %w", err)
	}
	return countries, nil
}
func (e *ExternalApi) FetchExchangeRate() (map[string]float64, error) {
	url := "https://open.er-api.com/v6/latest/USD"
	resp, err := e.httpclient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exchangereate API returned status %d", resp.StatusCode)
	}
	var exRate models.ExchangeRateResponse
	if err = json.NewDecoder(resp.Body).Decode(&exRate); err != nil {
		return nil, fmt.Errorf("failed to parse exchange rate JSON: %w", err)
	}
	if exRate.Result != "success" {
		return nil, fmt.Errorf("exchange rate API returned unsuccessful result")
	}
	return exRate.Rates, nil
}
