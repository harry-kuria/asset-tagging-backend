package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var jwtSecret string

func main() {
	// Load environment variables
	host := os.Getenv("HOST")
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	database := os.Getenv("DB")
	jwtSecret = os.Getenv("JWT_SECRET")
	port := os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}

	// Database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", username, password, host, database)
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to database successfully")

	// Initialize Gin router
	r := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// Serve static files
	r.Static("/files", "./files")
	r.Static("/assetLogos", "./assetLogos")

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Asset Tagging API is running",
		})
	})

	// Database test endpoint
	r.GET("/api/db-test", func(c *gin.Context) {
		// Check if database connection is available
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database connection not available",
			})
			return
		}

		// Test database connection
		err := db.Ping()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database connection failed: " + err.Error(),
			})
			return
		}

		// Check if assets table exists
		var tableExists int
		err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'assets'").Scan(&tableExists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check table existence: " + err.Error(),
			})
			return
		}

		if tableExists == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Assets table does not exist",
			})
			return
		}

		// Check table structure
		rows, err := db.Query("DESCRIBE assets")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to describe table: " + err.Error(),
			})
			return
		}
		defer rows.Close()

		var columns []gin.H
		for rows.Next() {
			var field, typ, null, key, defaultVal, extra sql.NullString
			err := rows.Scan(&field, &typ, &null, &key, &defaultVal, &extra)
			if err != nil {
				continue
			}
			columns = append(columns, gin.H{
				"field":   field.String,
				"type":    typ.String,
				"null":    null.String,
				"key":     key.String,
				"default": defaultVal.String,
				"extra":   extra.String,
			})
		}

		// Count total assets
		var assetCount int
		err = db.QueryRow("SELECT COUNT(*) FROM assets").Scan(&assetCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to count assets: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       "Database connected",
			"table_exists": tableExists > 0,
			"asset_count":  assetCount,
			"columns":      columns,
		})
	})

	// Public routes (no authentication required)
	public := r.Group("/api")
	{
		public.POST("/login", loginHandler)
		public.POST("/register/company", registerCompanyHandler)
		public.POST("/companies", createCompanyHandler) // New company creation endpoint
		
		// Reference data (public access)
		public.GET("/categories", getCategoriesHandler)
		public.GET("/institutions", getInstitutionsHandler)
		public.GET("/departments", getDepartmentsHandler)
		public.GET("/functional-areas", getFunctionalAreasHandler)
		public.GET("/manufacturers", getManufacturersHandler)
	}

	// Backward-compatible alias (legacy clients hitting /create-account)
	r.POST("/create-account", registerCompanyHandler)

	// Legacy endpoints (public access for frontend compatibility)
	r.POST("/addAsset", authMiddleware(), checkTrialStatusMiddleware(), addAssetHandler) // Legacy endpoint with auth and trial check

	// Protected routes (authentication required)
	protected := r.Group("/api")
	protected.Use(authMiddleware())
	{
		// Trial management (no trial check required)
		protected.GET("/trial/status", getTrialStatusHandler)
		protected.GET("/trial/plans", getPaymentPlansHandler)
		protected.POST("/trial/payment", initiatePaymentHandler)
		protected.POST("/payment/webhook", paymentWebhookHandler)

		// Company management
		protected.GET("/company", getCompanyHandler)
		protected.PUT("/company", updateCompanyHandler)
		protected.GET("/companies", listCompaniesHandler) // Admin only

		// User management
		protected.GET("/users", getUsersHandler)
		protected.GET("/users/:id", getUserHandler)
		protected.POST("/users", addUserHandler)
		protected.PUT("/users/:id", updateUserHandler)
		protected.DELETE("/users/:id", deleteUserHandler)

		// Asset management (requires active trial/subscription)
		assetRoutes := protected.Group("")
		assetRoutes.Use(checkTrialStatusMiddleware())
		{
			assetRoutes.GET("/assets", getAssetsHandler)
			assetRoutes.GET("/assets/:id", getAssetDetailsHandler)
			assetRoutes.POST("/assets", addAssetHandler)
			assetRoutes.POST("/assets/multiple", addMultipleAssetsHandler)
			assetRoutes.PUT("/assets/:id", updateAssetHandler)
			assetRoutes.DELETE("/assets/:id", deleteAssetHandler)
			assetRoutes.POST("/assets/search", searchAssetsHandler)

			// Asset categories (protected - for management)
			assetRoutes.POST("/categories", addCategoryHandler)
			assetRoutes.PUT("/categories/:id", updateCategoryHandler)
			assetRoutes.DELETE("/categories/:id", deleteCategoryHandler)

			// Barcode generation
			assetRoutes.POST("/barcodes", generateBarcodesHandler)
			assetRoutes.POST("/barcodes/institution", generateBarcodesByInstitutionHandler)
			assetRoutes.POST("/barcodes/institution-department", generateBarcodesByInstitutionAndDepartmentHandler)

			// Reports
			assetRoutes.POST("/reports", generateReportHandler)
			assetRoutes.GET("/generateReport", generateReportHandler) // Legacy GET endpoint
			assetRoutes.POST("/fetchAssetsByInstitution", fetchAssetsByInstitutionHandler) // For Excel reports
			assetRoutes.POST("/reports/assets", generateAssetReportHandler)
			assetRoutes.POST("/reports/invoice", generateInvoiceHandler)
			assetRoutes.GET("/reports/download/:filename", downloadHandler)

			// Dashboard
			assetRoutes.GET("/dashboard/stats", getDashboardStatsHandler)
			assetRoutes.GET("/dashboard/diagnostics", getDashboardDiagnosticsHandler) // Diagnostic endpoint for debugging
		}

		// Auth
		protected.POST("/logout", logoutHandler)
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
} 