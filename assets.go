package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// getAssetsHandler returns all assets
func getAssetsHandler(c *gin.Context) {
	rows, err := db.Query("SELECT id, assetName, assetType, institutionName, department, functionalArea, manufacturer, modelNumber, serialNumber, location, status, purchaseDate, purchasePrice, logo, createdAt, updatedAt FROM assets")
	if err != nil {
		log.Printf("Error fetching assets: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}
	defer rows.Close()

	var assets []gin.H
	for rows.Next() {
		var asset Asset
		err := rows.Scan(
			&asset.ID, &asset.AssetName, &asset.AssetType, &asset.InstitutionName, &asset.Department,
			&asset.FunctionalArea, &asset.Manufacturer, &asset.ModelNumber, &asset.SerialNumber,
			&asset.Location, &asset.Status, &asset.PurchaseDate, &asset.PurchasePrice,
			&asset.Logo, &asset.CreatedAt, &asset.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning asset: %v", err)
			continue
		}

		assets = append(assets, gin.H{
			"id":              asset.ID,
			"assetName":       asset.AssetName,
			"assetType":       asset.AssetType,
			"institutionName": asset.InstitutionName,
			"department":      asset.Department,
			"functionalArea":  asset.FunctionalArea,
			"manufacturer":    asset.Manufacturer,
			"modelNumber":     asset.ModelNumber,
			"serialNumber":    asset.SerialNumber,
			"location":        asset.Location,
			"status":          asset.Status,
			"purchaseDate":    asset.PurchaseDate.Format("2006-01-02"),
			"purchasePrice":   asset.PurchasePrice,
			"logo":            asset.Logo,
			"createdAt":       asset.CreatedAt.Format("2006-01-02 15:04:05"),
			"updatedAt":       asset.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, assets)
}

// addAssetHandler creates a new asset
func addAssetHandler(c *gin.Context) {
	var req AssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	// Parse purchase date
	var purchaseDate time.Time
	var err error
	if req.PurchaseDate != "" {
		purchaseDate, err = time.Parse("2006-01-02", req.PurchaseDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "Invalid purchase date format",
			})
			return
		}
	}

	// Insert the new asset
	result, err := db.Exec(`
		INSERT INTO assets (assetName, assetType, institutionName, department, functionalArea, 
		manufacturer, modelNumber, serialNumber, location, status, purchaseDate, purchasePrice, 
		logo, createdAt, updatedAt) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.AssetName, req.AssetType, req.InstitutionName, req.Department, req.FunctionalArea,
		req.Manufacturer, req.ModelNumber, req.SerialNumber, req.Location, req.Status,
		purchaseDate, req.PurchasePrice, "", time.Now(), time.Now())

	if err != nil {
		log.Printf("Error adding asset: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	assetID, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Asset added successfully",
		Data: gin.H{
			"assetId": assetID,
		},
	})
}

// updateAssetHandler updates an existing asset
func updateAssetHandler(c *gin.Context) {
	id := c.Param("id")
	assetID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid asset ID",
		})
		return
	}

	var req AssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	// Parse purchase date
	var purchaseDate time.Time
	if req.PurchaseDate != "" {
		purchaseDate, err = time.Parse("2006-01-02", req.PurchaseDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "Invalid purchase date format",
			})
			return
		}
	}

	// Update the asset
	_, err = db.Exec(`
		UPDATE assets SET assetName = ?, assetType = ?, institutionName = ?, department = ?, 
		functionalArea = ?, manufacturer = ?, modelNumber = ?, serialNumber = ?, location = ?, 
		status = ?, purchaseDate = ?, purchasePrice = ?, updatedAt = ? WHERE id = ?`,
		req.AssetName, req.AssetType, req.InstitutionName, req.Department, req.FunctionalArea,
		req.Manufacturer, req.ModelNumber, req.SerialNumber, req.Location, req.Status,
		purchaseDate, req.PurchasePrice, time.Now(), assetID)

	if err != nil {
		log.Printf("Error updating asset: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Asset updated successfully",
	})
}

// deleteAssetHandler deletes an asset
func deleteAssetHandler(c *gin.Context) {
	id := c.Param("id")
	assetID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid asset ID",
		})
		return
	}

	_, err = db.Exec("DELETE FROM assets WHERE id = ?", assetID)
	if err != nil {
		log.Printf("Error deleting asset: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Asset deleted successfully",
	})
}

// searchAssetsHandler searches for assets
func searchAssetsHandler(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Search query is required",
		})
		return
	}

	searchQuery := "%" + query + "%"
	rows, err := db.Query(`
		SELECT id, assetName, assetType, institutionName, department, functionalArea, 
		manufacturer, modelNumber, serialNumber, location, status, purchaseDate, 
		purchasePrice, logo, createdAt, updatedAt 
		FROM assets 
		WHERE assetName LIKE ? OR assetType LIKE ? OR institutionName LIKE ? OR 
		department LIKE ? OR manufacturer LIKE ? OR modelNumber LIKE ? OR 
		serialNumber LIKE ? OR location LIKE ?`,
		searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)

	if err != nil {
		log.Printf("Error searching assets: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}
	defer rows.Close()

	var assets []gin.H
	for rows.Next() {
		var asset Asset
		err := rows.Scan(
			&asset.ID, &asset.AssetName, &asset.AssetType, &asset.InstitutionName, &asset.Department,
			&asset.FunctionalArea, &asset.Manufacturer, &asset.ModelNumber, &asset.SerialNumber,
			&asset.Location, &asset.Status, &asset.PurchaseDate, &asset.PurchasePrice,
			&asset.Logo, &asset.CreatedAt, &asset.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning asset: %v", err)
			continue
		}

		assets = append(assets, gin.H{
			"id":              asset.ID,
			"assetName":       asset.AssetName,
			"assetType":       asset.AssetType,
			"institutionName": asset.InstitutionName,
			"department":      asset.Department,
			"functionalArea":  asset.FunctionalArea,
			"manufacturer":    asset.Manufacturer,
			"modelNumber":     asset.ModelNumber,
			"serialNumber":    asset.SerialNumber,
			"location":        asset.Location,
			"status":          asset.Status,
			"purchaseDate":    asset.PurchaseDate.Format("2006-01-02"),
			"purchasePrice":   asset.PurchasePrice,
			"logo":            asset.Logo,
			"createdAt":       asset.CreatedAt.Format("2006-01-02 15:04:05"),
			"updatedAt":       asset.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, assets)
}

// getAssetDetailsHandler returns details of a specific asset
func getAssetDetailsHandler(c *gin.Context) {
	assetIDStr := c.Query("assetId")
	assetID, err := strconv.Atoi(assetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid asset ID",
		})
		return
	}

	var asset Asset
	err = db.QueryRow(`
		SELECT id, assetName, assetType, institutionName, department, functionalArea, 
		manufacturer, modelNumber, serialNumber, location, status, purchaseDate, 
		purchasePrice, logo, createdAt, updatedAt 
		FROM assets WHERE id = ?`, assetID).
		Scan(&asset.ID, &asset.AssetName, &asset.AssetType, &asset.InstitutionName, &asset.Department,
			&asset.FunctionalArea, &asset.Manufacturer, &asset.ModelNumber, &asset.SerialNumber,
			&asset.Location, &asset.Status, &asset.PurchaseDate, &asset.PurchasePrice,
			&asset.Logo, &asset.CreatedAt, &asset.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Error:   "Asset not found",
			})
			return
		}
		log.Printf("Error fetching asset details: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              asset.ID,
		"assetName":       asset.AssetName,
		"assetType":       asset.AssetType,
		"institutionName": asset.InstitutionName,
		"department":      asset.Department,
		"functionalArea":  asset.FunctionalArea,
		"manufacturer":    asset.Manufacturer,
		"modelNumber":     asset.ModelNumber,
		"serialNumber":    asset.SerialNumber,
		"location":        asset.Location,
		"status":          asset.Status,
		"purchaseDate":    asset.PurchaseDate.Format("2006-01-02"),
		"purchasePrice":   asset.PurchasePrice,
		"logo":            asset.Logo,
		"createdAt":       asset.CreatedAt.Format("2006-01-02 15:04:05"),
		"updatedAt":       asset.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// addMultipleAssetsHandler adds multiple assets
func addMultipleAssetsHandler(c *gin.Context) {
	assetType := c.Param("assetType")
	var req MultipleAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	var assetIDs []int64
	for _, assetReq := range req.Assets {
		// Parse purchase date
		var purchaseDate time.Time
		var err error
		if assetReq.PurchaseDate != "" {
			purchaseDate, err = time.Parse("2006-01-02", assetReq.PurchaseDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, APIResponse{
					Success: false,
					Error:   "Invalid purchase date format",
				})
				return
			}
		}

		// Insert the asset
		result, err := db.Exec(`
			INSERT INTO assets (assetName, assetType, institutionName, department, functionalArea, 
			manufacturer, modelNumber, serialNumber, location, status, purchaseDate, purchasePrice, 
			logo, createdAt, updatedAt) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			assetReq.AssetName, assetType, assetReq.InstitutionName, assetReq.Department, assetReq.FunctionalArea,
			assetReq.Manufacturer, assetReq.ModelNumber, assetReq.SerialNumber, assetReq.Location, assetReq.Status,
			purchaseDate, assetReq.PurchasePrice, "", time.Now(), time.Now())

		if err != nil {
			log.Printf("Error adding asset: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   "Internal Server Error",
			})
			return
		}

		assetID, _ := result.LastInsertId()
		assetIDs = append(assetIDs, assetID)
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully added %d assets", len(req.Assets)),
		Data: gin.H{
			"assetIds": assetIDs,
		},
	})
}

// getInstitutionsHandler returns all unique institutions
func getInstitutionsHandler(c *gin.Context) {
	// Check if database connection is available
	if db == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database connection not available",
		})
		return
	}

	// Test database connection
	err := db.Ping()
	if err != nil {
		log.Printf("Database ping failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database connection failed",
		})
		return
	}

	// Check if assets table exists
	var tableExists int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'assets'").Scan(&tableExists)
	if err != nil {
		log.Printf("Error checking table existence: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to check table existence",
		})
		return
	}

	if tableExists == 0 {
		// Table doesn't exist, return empty array
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data:    []string{},
		})
		return
	}

	rows, err := db.Query("SELECT DISTINCT institutionName FROM assets WHERE institutionName IS NOT NULL AND institutionName != ''")
	if err != nil {
		log.Printf("Error fetching institutions: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to fetch institutions: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var institutions []string
	for rows.Next() {
		var institution string
		err := rows.Scan(&institution)
		if err != nil {
			log.Printf("Error scanning institution: %v", err)
			continue
		}
		if institution != "" {
			institutions = append(institutions, institution)
		}
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating institutions: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Error processing institutions",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    institutions,
	})
}

// getDepartmentsHandler returns all unique departments
func getDepartmentsHandler(c *gin.Context) {
	// Check if database connection is available
	if db == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database connection not available",
		})
		return
	}

	// Test database connection
	err := db.Ping()
	if err != nil {
		log.Printf("Database ping failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database connection failed",
		})
		return
	}

	// Check if assets table exists
	var tableExists int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'assets'").Scan(&tableExists)
	if err != nil {
		log.Printf("Error checking table existence: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to check table existence",
		})
		return
	}

	if tableExists == 0 {
		// Table doesn't exist, return empty array
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data:    []string{},
		})
		return
	}

	rows, err := db.Query("SELECT DISTINCT department FROM assets WHERE department IS NOT NULL AND department != ''")
	if err != nil {
		log.Printf("Error fetching departments: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to fetch departments: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var departments []string
	for rows.Next() {
		var department string
		err := rows.Scan(&department)
		if err != nil {
			log.Printf("Error scanning department: %v", err)
			continue
		}
		if department != "" {
			departments = append(departments, department)
		}
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating departments: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Error processing departments",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    departments,
	})
}

// getFunctionalAreasHandler returns all unique functional areas
func getFunctionalAreasHandler(c *gin.Context) {
	// Check if database connection is available
	if db == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database connection not available",
		})
		return
	}

	// Test database connection
	err := db.Ping()
	if err != nil {
		log.Printf("Database ping failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database connection failed",
		})
		return
	}

	// Check if assets table exists
	var tableExists int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'assets'").Scan(&tableExists)
	if err != nil {
		log.Printf("Error checking table existence: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to check table existence",
		})
		return
	}

	if tableExists == 0 {
		// Table doesn't exist, return empty array
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data:    []string{},
		})
		return
	}

	rows, err := db.Query("SELECT DISTINCT functionalArea FROM assets WHERE functionalArea IS NOT NULL AND functionalArea != ''")
	if err != nil {
		log.Printf("Error fetching functional areas: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to fetch functional areas: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var functionalAreas []string
	for rows.Next() {
		var functionalArea string
		err := rows.Scan(&functionalArea)
		if err != nil {
			log.Printf("Error scanning functional area: %v", err)
			continue
		}
		if functionalArea != "" {
			functionalAreas = append(functionalAreas, functionalArea)
		}
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating functional areas: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Error processing functional areas",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    functionalAreas,
	})
}

// getManufacturersHandler returns all unique manufacturers
func getManufacturersHandler(c *gin.Context) {
	// Check if database connection is available
	if db == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database connection not available",
		})
		return
	}

	// Test database connection
	err := db.Ping()
	if err != nil {
		log.Printf("Database ping failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database connection failed",
		})
		return
	}

	// Check if assets table exists
	var tableExists int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'assets'").Scan(&tableExists)
	if err != nil {
		log.Printf("Error checking table existence: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to check table existence",
		})
		return
	}

	if tableExists == 0 {
		// Table doesn't exist, return empty array
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data:    []string{},
		})
		return
	}

	rows, err := db.Query("SELECT DISTINCT manufacturer FROM assets WHERE manufacturer IS NOT NULL AND manufacturer != ''")
	if err != nil {
		log.Printf("Error fetching manufacturers: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to fetch manufacturers: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var manufacturers []string
	for rows.Next() {
		var manufacturer string
		err := rows.Scan(&manufacturer)
		if err != nil {
			log.Printf("Error scanning manufacturer: %v", err)
			continue
		}
		if manufacturer != "" {
			manufacturers = append(manufacturers, manufacturer)
		}
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating manufacturers: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Error processing manufacturers",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    manufacturers,
	})
} 