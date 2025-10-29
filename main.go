package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	db "github.com/franzego/stage02/db/sqlc"
	"github.com/franzego/stage02/internal"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env only in local development
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found (this is normal on Railway)")
	}

	// Build DSN (Railway-compatible)
	dsn := buildDSN()

	// Connect to database
	dbconn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbconn.Close()

	// Test connection
	if err := dbconn.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Database connected")

	// Initialize queries
	queries := db.New(dbconn)
	handle := internal.NewCountryHandler(queries)

	// Setup Gin
	r := gin.Default()

	// Routes
	r.POST("/countries/refresh", handle.RefreshCountries)
	r.GET("/countries", handle.GetAllCountries)
	r.GET("/countries/:name", handle.GetCountryName)
	r.DELETE("/countries/:name", handle.DeleteCountryName)
	r.GET("/status", handle.GetStatus)
	r.GET("/countries/image", handle.GetImage)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Ensure cache directory exists
	cacheDir := getEnv("CACHE_DIR", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("Failed to create cache dir: %v", err)
	}

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("🚀 Server starting on port %s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// buildDSN creates Railway-compatible connection string
func buildDSN() string {
	if mysqlURL := os.Getenv("MYSQL_URL"); mysqlURL != "" {
		u, err := url.Parse(mysqlURL)
		if err != nil {
			log.Fatalf("Invalid MYSQL_URL: %v", err)
		}

		user := u.User.Username()
		password, _ := u.User.Password()
		host := u.Host
		dbName := strings.TrimPrefix(u.Path, "/")

		return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, password, host, dbName)
	}

	// fallback
	user := getEnv("MYSQLUSER", os.Getenv("DB_USER"))
	password := getEnv("MYSQLPASSWORD", os.Getenv("DB_PASSWORD"))
	host := getEnv("MYSQLHOST", os.Getenv("DB_HOST"))
	port := getEnv("MYSQLPORT", os.Getenv("DB_PORT"))
	database := getEnv("MYSQLDATABASE", os.Getenv("DB_NAME"))

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, database)
}

// getEnv gets env variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback

}
