package models

type CountryData struct {
	Name        string     `json:"name"`
	Capital     string     `json:"capital"`
	Region      string     `json:"region"`
	Population  int64      `json:"population"`
	Flag        string     `json:"flag"`
	Currencies  []Currency `json:"currencies"`
	Independent bool       `json:"independent"`
}
type Currency struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	SymbolUrl string `json:"symbol"`
}
type ExchangeRateResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}
type ProcessedCountry struct {
	Name         string
	Capital      string
	Region       string
	Population   int64
	CurrencyCode *string  // nullable
	ExchangeRate *float64 // nullable
	EstimatedGDP *float64 // nullable
	FlagURL      string
}
type CountryResponse struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Capital         *string  `json:"capital,omitempty"`
	Region          *string  `json:"region,omitempty"`
	Population      int64    `json:"population"`
	CurrencyCode    *string  `json:"currency_code,omitempty"`
	ExchangeRate    *float64 `json:"exchange_rate,omitempty"`
	EstimatedGDP    *float64 `json:"estimated_gdp,omitempty"`
	FlagURL         *string  `json:"flag_url,omitempty"`
	LastRefreshedAt string   `json:"last_refreshed_at,omitempty"`
}
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details"`
}
type MessageResponse struct {
	Message string `json:"message"`
}
type StatusResponse struct {
	TotalCountries  int64       `json:"total_countries"`
	LastRefreshedAt interface{} `json:"last_refreshed_at"`
}
