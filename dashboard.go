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
	err := db.QueryRow("SELECT COUNT(*) FROM assets WHERE company_id = ?", companyID).Scan(&totalAssets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get total assets: " + err.Error(),
		})
		return
	}

	// Get active assets
	var activeAssets int
	err = db.QueryRow("SELECT COUNT(*) FROM assets WHERE company_id = ? AND status = 'Active'", companyID).Scan(&activeAssets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get active assets: " + err.Error(),
		})
		return
	}

	// Get total users - handle potential column name mismatches
	var totalUsers int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE company_id = ? AND is_active = true", companyID).Scan(&totalUsers)
	if err != nil {
		// If users table doesn't exist or has different structure, default to 0
		totalUsers = 0
	}

	// Get total value
	var totalValue float64
	err = db.QueryRow("SELECT COALESCE(SUM(purchase_price), 0) FROM assets WHERE company_id = ? AND purchase_price IS NOT NULL", companyID).Scan(&totalValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get total value: " + err.Error(),
		})
		return
	}

	// Get total barcodes (assets with barcode or QR code)
	var totalBarcodes int
	err = db.QueryRow("SELECT COUNT(*) FROM assets WHERE company_id = ? AND (barcode IS NOT NULL OR qr_code IS NOT NULL)", companyID).Scan(&totalBarcodes)
	if err != nil {
		// If barcode columns don't exist, default to 0
		totalBarcodes = 0
	}

	// Get scanned barcodes (assets that have been scanned - for now, we'll use assets with recent activity)
	var scannedBarcodes int
	err = db.QueryRow("SELECT COUNT(*) FROM assets WHERE company_id = ? AND updated_at > DATE_SUB(NOW(), INTERVAL 30 DAY)", companyID).Scan(&scannedBarcodes)
	if err != nil {
		// If there's an error, default to 0
		scannedBarcodes = 0
	}

	// Get assets by status
	rows, err := db.Query("SELECT status, COUNT(*) FROM assets WHERE company_id = ? GROUP BY status", companyID)
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
	rows, err = db.Query("SELECT asset_type, COUNT(*) FROM assets WHERE company_id = ? AND asset_type IS NOT NULL GROUP BY asset_type", companyID)
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
		SELECT id, company_id, asset_name, asset_type, institution_name, department, 
		functional_area, manufacturer, model_number, serial_number, location, status, 
		purchase_date, purchase_price, created_at, updated_at
		FROM assets 
		WHERE company_id = ? 
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
		SELECT id, company_id, asset_id, maintenance_type, description, cost,
		performed_by, performed_at, next_maintenance_date, created_by, created_at
		FROM asset_maintenance 
		WHERE company_id = ? 
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
		TotalBarcodes:     totalBarcodes,
		ScannedBarcodes:   scannedBarcodes,
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

// getDashboardDiagnosticsHandler provides diagnostic information for debugging company/asset issues
func getDashboardDiagnosticsHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)
	user := getCurrentUser(c)

	// Get company information
	var company Company
	err := db.QueryRow(`
		SELECT id, company_name, company_code, email, subscription_plan, is_active, trial_ends_at, created_at, updated_at
		FROM companies WHERE id = ?
	`, companyID).Scan(
		&company.ID, &company.CompanyName, &company.CompanyCode, &company.Email,
		&company.SubscriptionPlan, &company.IsActive, &company.TrialEndsAt,
		&company.CreatedAt, &company.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get company info: " + err.Error(),
		})
		return
	}

	// Get total assets for this company
	var totalAssets int
	err = db.QueryRow("SELECT COUNT(*) FROM assets WHERE company_id = ?", companyID).Scan(&totalAssets)
	if err != nil {
		totalAssets = 0
	}

	// Get total users for this company
	var totalUsers int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE company_id = ? AND is_active = true", companyID).Scan(&totalUsers)
	if err != nil {
		totalUsers = 0
	}

	// Get all companies
	rows, err := db.Query("SELECT id, company_name, company_code FROM companies ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get companies: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var companies []map[string]interface{}
	for rows.Next() {
		var id int
		var name, code string
		err := rows.Scan(&id, &name, &code)
		if err != nil {
			continue
		}
		companies = append(companies, map[string]interface{}{
			"id": id, "name": name, "code": code,
		})
	}

	// Get assets by company
	rows, err = db.Query("SELECT company_id, COUNT(*) FROM assets GROUP BY company_id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get assets by company: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	assetsByCompany := make(map[int]int)
	for rows.Next() {
		var companyID, count int
		err := rows.Scan(&companyID, &count)
		if err != nil {
			continue
		}
		assetsByCompany[companyID] = count
	}

	// Get sample assets for debugging
	rows, err = db.Query("SELECT id, company_id, asset_name, created_at FROM assets ORDER BY created_at DESC LIMIT 5")
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get sample assets: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var sampleAssets []map[string]interface{}
	for rows.Next() {
		var id, assetCompanyID int
		var assetName, createdAt string
		err := rows.Scan(&id, &assetCompanyID, &assetName, &createdAt)
		if err != nil {
			continue
		}
		sampleAssets = append(sampleAssets, map[string]interface{}{
			"id": id, "company_id": assetCompanyID, "asset_name": assetName, "created_at": createdAt,
		})
	}

	diagnostics := map[string]interface{}{
		"current_user": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"company_id": user.CompanyID,
		},
		"current_company": company,
		"current_company_stats": map[string]interface{}{
			"total_assets": totalAssets,
			"total_users":  totalUsers,
		},
		"all_companies":     companies,
		"assets_by_company": assetsByCompany,
		"sample_assets":     sampleAssets,
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    diagnostics,
	})
} 