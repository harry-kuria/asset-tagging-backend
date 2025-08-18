package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getDashboardStatsHandler returns dashboard statistics for the current company
func getDashboardStatsHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)

	// Get total assets
	var totalAssets int
	err := db.QueryRow("SELECT COUNT(*) FROM assets WHERE companyId = ?", companyID).Scan(&totalAssets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get total assets: " + err.Error(),
		})
		return
	}

	// Get active assets
	var activeAssets int
	err = db.QueryRow("SELECT COUNT(*) FROM assets WHERE companyId = ? AND status = 'Active'", companyID).Scan(&activeAssets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get active assets: " + err.Error(),
		})
		return
	}

	// Get total users - handle potential column name mismatches
	var totalUsers int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE companyId = ? AND is_active = true", companyID).Scan(&totalUsers)
	if err != nil {
		// If users table doesn't exist or has different structure, default to 0
		totalUsers = 0
	}

	// Get total value
	var totalValue float64
	err = db.QueryRow("SELECT COALESCE(SUM(purchase_price), 0) FROM assets WHERE companyId = ? AND purchase_price IS NOT NULL", companyID).Scan(&totalValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get total value: " + err.Error(),
		})
		return
	}

	// Get assets by status
	rows, err := db.Query("SELECT status, COUNT(*) FROM assets WHERE companyId = ? GROUP BY status", companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get assets by status: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	assetsByStatus := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		err := rows.Scan(&status, &count)
		if err != nil {
			continue
		}
		assetsByStatus[status] = count
	}

	// Get assets by type
	rows, err = db.Query("SELECT asset_type, COUNT(*) FROM assets WHERE companyId = ? AND asset_type IS NOT NULL GROUP BY asset_type", companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get assets by type: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	assetsByType := make(map[string]int)
	for rows.Next() {
		var assetType string
		var count int
		err := rows.Scan(&assetType, &count)
		if err != nil {
			continue
		}
		assetsByType[assetType] = count
	}

	// Get recent assets (last 10)
	rows, err = db.Query(`
		SELECT id, companyId, asset_name, asset_type, institution_name, department, 
		functional_area, manufacturer, model_number, serial_number, location, status, 
		purchase_date, purchase_price, created_at, updated_at
		FROM assets 
		WHERE companyId = ? 
		ORDER BY created_at DESC 
		LIMIT 10
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get recent assets: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var recentAssets []Asset
	for rows.Next() {
		var asset Asset
		err := rows.Scan(
			&asset.ID, &asset.CompanyID, &asset.AssetName, &asset.AssetType,
			&asset.InstitutionName, &asset.Department, &asset.FunctionalArea,
			&asset.Manufacturer, &asset.ModelNumber, &asset.SerialNumber,
			&asset.Location, &asset.Status, &asset.PurchaseDate, &asset.PurchasePrice,
			&asset.CreatedAt, &asset.UpdatedAt,
		)
		if err != nil {
			continue
		}
		recentAssets = append(recentAssets, asset)
	}

	// Get recent maintenance (last 5) - handle case where table might not exist
	var recentMaintenance []AssetMaintenance
	rows, err = db.Query(`
		SELECT id, companyId, asset_id, maintenance_type, description, cost,
		performed_by, performed_at, next_maintenance_date, created_by, created_at
		FROM asset_maintenance 
		WHERE companyId = ? 
		ORDER BY performed_at DESC 
		LIMIT 5
	`, companyID)
	if err != nil {
		// If table doesn't exist, just continue with empty maintenance data
		// This is not a critical error for dashboard functionality
		recentMaintenance = []AssetMaintenance{}
	} else {
		defer rows.Close()
		
		for rows.Next() {
			var maintenance AssetMaintenance
			err := rows.Scan(
				&maintenance.ID, &maintenance.CompanyID, &maintenance.AssetID,
				&maintenance.MaintenanceType, &maintenance.Description, &maintenance.Cost,
				&maintenance.PerformedBy, &maintenance.PerformedAt, &maintenance.NextMaintenanceDate,
				&maintenance.CreatedBy, &maintenance.CreatedAt,
			)
			if err != nil {
				continue
			}
			recentMaintenance = append(recentMaintenance, maintenance)
		}
	}

	stats := DashboardStats{
		TotalAssets:       totalAssets,
		ActiveAssets:      activeAssets,
		TotalUsers:        totalUsers,
		TotalValue:        totalValue,
		AssetsByStatus:    assetsByStatus,
		AssetsByType:      assetsByType,
		RecentAssets:      recentAssets,
		RecentMaintenance: recentMaintenance,
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    stats,
	})
} 