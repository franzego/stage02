package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	db "github.com/franzego/stage02/db/sqlc"
	"github.com/franzego/stage02/internal"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file")
	}

	// Build database connection string
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") +
		"@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") +
		")/" + os.Getenv("DB_NAME") + "?parseTime=true"

	// Connect to database
	dbconn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbconn.Close()

	// Test the connection
	if err := dbconn.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established")

	// Initialize SQLC queries
	queries := db.New(dbconn)
	handle := internal.NewCountryHandler(queries)

	// Setup Gin router
	r := gin.Default()
	// routes
	r.POST("/countries/refresh", handle.RefreshCountries)
	r.GET("/countries", handle.GetAllCountries)
	r.GET("/countries/:name", handle.GetCountryName)
	r.DELETE("/countries/:name", handle.DeleteCountryName)
	r.GET("/status", handle.GetStatus)
	r.GET("/countries/image", handle.GetImage)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// ensure cache dir exists at startup
	cacheDir := os.Getenv("CACHE_DIR")
	if cacheDir == "" {
		cacheDir = "cache"
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("Failed to create cache dir: %v", err)
	}

	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		if err := r.Run("0.0.0.0:" + port); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	log.Println("Shutting down gracefully...")

}
