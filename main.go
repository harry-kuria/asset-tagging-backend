package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database connection
	initDB()

	// Set up Gin router
	router := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	// Serve static files
	router.Static("/assetLogos", "./assetLogos")

	// API routes
	api := router.Group("/api")
	{
		// Public routes (no authentication required)
		api.POST("/login", loginHandler)
		api.POST("/create-account", createAccountHandler)

		// Protected routes (authentication required)
		protected := api.Group("")
		protected.Use(authMiddleware)
		{
			// User management routes
			protected.GET("/users", getUsersHandler)
			protected.GET("/users/:id", getUserHandler)
			protected.POST("/addUser", addUserHandler)
			protected.PUT("/users/:id", updateUserHandler)
			protected.DELETE("/users/:id", deleteUserHandler)

			// Asset management routes
			protected.GET("/assets", getAssetsHandler)
			protected.POST("/addAsset", addAssetHandler)
			protected.PUT("/assets/:id", updateAssetHandler)
			protected.DELETE("/assets/:id", deleteAssetHandler)
			protected.GET("/searchAssets", searchAssetsHandler)
			protected.GET("/getAssetDetails", getAssetDetailsHandler)

			// Multiple assets
			protected.POST("/addMultipleAssets/:assetType", addMultipleAssetsHandler)

			// Reference data routes
			protected.GET("/institutions", getInstitutionsHandler)
			protected.GET("/departments", getDepartmentsHandler)
			protected.GET("/functionalAreas", getFunctionalAreasHandler)
			protected.GET("/categories", getCategoriesHandler)
			protected.GET("/manufacturers", getManufacturersHandler)

			// Barcode generation routes
			protected.POST("/generateBarcodes", generateBarcodesHandler)
			protected.POST("/generateBarcodesByInstitution", generateBarcodesByInstitutionHandler)
			protected.POST("/generateBarcodesByInstitutionAndDepartment", generateBarcodesByInstitutionAndDepartmentHandler)

			// Report routes
			protected.POST("/generateAssetReport", generateAssetReportHandler)
			protected.GET("/generateReport", generateReportHandler)
			protected.POST("/generate_invoice", generateInvoiceHandler)
			protected.GET("/download", downloadHandler)

			// Asset fetching routes
			protected.POST("/fetchAssetsByInstitution", fetchAssetsByInstitutionHandler)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func initDB() {
	host := os.Getenv("HOST")
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	database := os.Getenv("DB")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", username, password, host, database)
	
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Connected to MySQL database")
} 