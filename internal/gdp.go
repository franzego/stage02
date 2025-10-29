package internal

import (
	"database/sql"
	"strconv"
)

// ParseNullStringFloat parses a sql.NullString containing a numeric value
// and returns a float64. Returns 0 if invalid or on parse error.
func ParseNullStringFloat(ns sql.NullString) float64 {
	if !ns.Valid {
		return 0
	}
	v, err := strconv.ParseFloat(ns.String, 64)
	if err != nil {
		return 0
	}
	return v
}
