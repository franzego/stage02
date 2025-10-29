package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	db "github.com/franzego/stage02/db/sqlc"
	"github.com/franzego/stage02/internal"
	"github.com/franzego/stage02/internal/database"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found (this is normal on Railway)")
	}
	dsn := buildDSN()
	dbconn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbconn.Close()
	if err := dbconn.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Database connected")

	if err := database.RunMigrations(dbconn); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	// Initialize queries
	queries := db.New(dbconn)
	handle := internal.NewCountryHandler(queries)
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
	cacheDir := getEnv("CACHE_DIR", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("Failed to create cache dir: %v", err)
	}
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// buildDSN creates Railway-compatible connection string
func buildDSN() string {
	if mysqlURL := os.Getenv("MYSQL_URL"); mysqlURL != "" {
		return mysqlURL
	}

	// Fallback: build from individual variables
	user := getEnv("MYSQLUSER", os.Getenv("DB_USER"))
	password := getEnv("MYSQLPASSWORD", os.Getenv("DB_PASSWORD"))
	host := getEnv("MYSQLHOST", os.Getenv("DB_HOST"))
	port := getEnv("MYSQLPORT", os.Getenv("DB_PORT"))
	database := getEnv("MYSQLDATABASE", os.Getenv("DB_NAME"))

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, password, host, port, database)
}

// getEnv gets env variable with fallback
// just iincase
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback

}
