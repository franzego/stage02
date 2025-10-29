package internal

import (
	"database/sql"
	"net/http"
	"os"
	"sort"
	"strings"

	db "github.com/franzego/stage02/db/sqlc"
	internal "github.com/franzego/stage02/internal/services"
	"github.com/franzego/stage02/models"
	"github.com/gin-gonic/gin"
)

type CountryHandler struct {
	service *internal.CountryService
	queries *db.Queries
}

func NewCountryHandler(queries *db.Queries) *CountryHandler {
	return &CountryHandler{
		service: internal.NewCountryService(queries),
		queries: queries,
	}
}

// POST /countries/refresh
func (h *CountryHandler) RefreshCountries(c *gin.Context) {
	if err := h.service.RefreshCountries(); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "External data source unavailable",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Countries refreshed successfully",
	})
}

// Get /countries
func (h *CountryHandler) GetAllCountries(c *gin.Context) {
	countries, err := h.service.GetAllCountries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "External data source not available",
			Details: err.Error(),
		})
		return
	}

	// filter by region
	region := c.Query("region")
	currency := c.Query("currency")
	sortBy := c.Query("sort")
	if region != "" {
		filtered := []db.Country{}
		for _, country := range countries {
			if country.Region.Valid && strings.EqualFold(country.Region.String, region) {
				filtered = append(filtered, country)
			}
		}
		countries = filtered
	}
	if currency != "" {
		filtered := []db.Country{}
		for _, country := range countries {
			if country.CurrencyCode.Valid && strings.EqualFold(country.CurrencyCode.String, currency) {
				filtered = append(filtered, country)
			}
		}
		countries = filtered
	}
	if sortBy == "gdp_desc" {
		sort.Slice(countries, func(i, j int) bool {
			return ParseNullStringFloat(countries[i].EstimatedGdp) > ParseNullStringFloat(countries[j].EstimatedGdp)
		})
	} else if sortBy == "gdp_asc" {
		sort.Slice(countries, func(i, j int) bool {
			return ParseNullStringFloat(countries[i].EstimatedGdp) < ParseNullStringFloat(countries[j].EstimatedGdp)
		})
	}

	// Map DB models to response models
	responses := make([]models.CountryResponse, 0, len(countries))
	for _, ct := range countries {
		responses = append(responses, h.mapCountryToResponse(ct))
	}
	c.JSON(http.StatusOK, responses)
}

// Get /countries/:name
func (h *CountryHandler) GetCountryName(c *gin.Context) {
	name := c.Param("name")
	country, err := h.service.GetCountryByName(name)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Country not found",
				Details: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Errror",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, h.mapCountryToResponse(country))

}

// Delete /countries/:name
func (h *CountryHandler) DeleteCountryName(c *gin.Context) {
	name := c.Param("name")
	if err := h.service.DeleteCountryByName(name); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Country not found",
				Details: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Details: err.Error(),
		})
		return

	}
	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Country successfully Deleted",
	})
}

// func to get status
func (h *CountryHandler) GetStatus(c *gin.Context) {
	count, err := h.service.GetTotalCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Details: err.Error(),
		})
		return
	}
	lastRefresh, err := h.service.GetRefreshTime()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Details: err.Error(),
		})
		return
	}
	response := models.StatusResponse{
		TotalCountries: count,
	}

	if lastRefresh.Valid {
		response.LastRefreshedAt = lastRefresh.Time.Format("2006-01-02T15:04:05Z")
	} else {
		response.LastRefreshedAt = "Never"
	}

	c.JSON(http.StatusOK, response)
}

// func to generate handler
func (h *CountryHandler) GetImage(c *gin.Context) {
	imagePath := "cache/summary.png"

	// Check if image exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Summary image not found",
		})
		return
	}
	c.File(imagePath)
}

// Helper function to map DB model to response model
func (h *CountryHandler) mapCountryToResponse(country db.Country) models.CountryResponse {
	response := models.CountryResponse{
		ID:         country.ID,
		Name:       country.Name,
		Population: country.Population,
	}

	if country.LastRefreshedAt.Valid {
		response.LastRefreshedAt = country.LastRefreshedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	if country.Capital.Valid {
		response.Capital = &country.Capital.String
	}

	if country.Region.Valid {
		response.Region = &country.Region.String
	}

	if country.CurrencyCode.Valid {
		response.CurrencyCode = &country.CurrencyCode.String
	}

	if country.ExchangeRate.Valid {
		v := ParseNullStringFloat(country.ExchangeRate)
		response.ExchangeRate = &v
	}

	if country.EstimatedGdp.Valid {
		v := ParseNullStringFloat(country.EstimatedGdp)
		response.EstimatedGDP = &v
	}

	if country.FlagUrl.Valid {
		response.FlagURL = &country.FlagUrl.String
	}

	return response
}
