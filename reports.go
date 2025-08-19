package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

// generateReportHandler generates a filtered report
func generateReportHandler(c *gin.Context) {
	// Get query parameters
	assetType := c.Query("assetType")
	location := c.Query("location")
	status := c.Query("status")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	manufacturer := c.Query("manufacturer")
	modelNumber := c.Query("modelNumber")
	institutionName := c.Query("institutionName")
	department := c.Query("department")
	functionalArea := c.Query("functionalArea")

	// Validate required parameters
	if institutionName == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Institution name is required",
		})
		return
	}

	// Build query dynamically
	query := "SELECT id, asset_name, asset_type, institution_name, department, functional_area, manufacturer, model_number, serial_number, location, status, purchase_date, purchase_price, created_at, updated_at FROM assets WHERE institution_name = ?"
	var args []interface{}
	args = append(args, institutionName)

	if assetType != "" && assetType != "All" {
		query += " AND asset_type = ?"
		args = append(args, assetType)
	}

	if location != "" && location != "All" {
		query += " AND location = ?"
		args = append(args, location)
	}

	if status != "" && status != "All" {
		query += " AND status = ?"
		args = append(args, status)
	}

	if startDate != "" {
		query += " AND purchase_date >= ?"
		args = append(args, startDate)
	}

	if endDate != "" {
		query += " AND purchase_date <= ?"
		args = append(args, endDate)
	}

	if manufacturer != "" && manufacturer != "All" {
		query += " AND manufacturer = ?"
		args = append(args, manufacturer)
	}

	if modelNumber != "" && modelNumber != "All" {
		query += " AND model_number = ?"
		args = append(args, modelNumber)
	}

	if department != "" && department != "All" {
		query += " AND department = ?"
		args = append(args, department)
	}

	if functionalArea != "" && functionalArea != "All" {
		query += " AND functional_area = ?"
		args = append(args, functionalArea)
	}

	// Add ordering
	query += " ORDER BY asset_name"

	log.Printf("Executing report query: %s with args: %v", query, args)

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error fetching assets for report: %v", err)
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
			&asset.CreatedAt, &asset.UpdatedAt)
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
			"marketValue":     asset.PurchasePrice, // Use purchase price as market value
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

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating assets: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Error processing report data",
		})
		return
	}

	log.Printf("Report generated successfully: %d assets found", len(assets))
	c.JSON(http.StatusOK, assets)
}

// generateAssetReportHandler generates a detailed asset report
func generateAssetReportHandler(c *gin.Context) {
	var req ReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	// Build query dynamically
	query := "SELECT id, assetName, assetType, institutionName, department, functionalArea, manufacturer, modelNumber, serialNumber, location, status, purchaseDate, purchasePrice, logo, createdAt, updatedAt FROM assets WHERE 1=1"
	var args []interface{}

	if req.AssetType != "" && req.AssetType != "All" {
		query += " AND assetType = ?"
		args = append(args, req.AssetType)
	}

	if req.Location != "" && req.Location != "All" {
		query += " AND location = ?"
		args = append(args, req.Location)
	}

	if req.Status != "" && req.Status != "All" {
		query += " AND status = ?"
		args = append(args, req.Status)
	}

	if req.StartDate != "" {
		query += " AND purchaseDate >= ?"
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		query += " AND purchaseDate <= ?"
		args = append(args, req.EndDate)
	}

	if len(req.Manufacturer) > 0 && req.Manufacturer[0] != "All" {
		placeholders := strings.Repeat("?,", len(req.Manufacturer))
		placeholders = placeholders[:len(placeholders)-1]
		query += fmt.Sprintf(" AND manufacturer IN (%s)", placeholders)
		for _, m := range req.Manufacturer {
			args = append(args, m)
		}
	}

	if req.ModelNumber != "" && req.ModelNumber != "All" {
		query += " AND modelNumber = ?"
		args = append(args, req.ModelNumber)
	}

	if req.InstitutionName != "" && req.InstitutionName != "All" {
		query += " AND institutionName = ?"
		args = append(args, req.InstitutionName)
	}

	if req.Department != "" && req.Department != "All" {
		query += " AND department = ?"
		args = append(args, req.Department)
	}

	if req.FunctionalArea != "" && req.FunctionalArea != "All" {
		query += " AND functionalArea = ?"
		args = append(args, req.FunctionalArea)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error fetching assets for report: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}
	defer rows.Close()

	var assets []Asset
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
		assets = append(assets, asset)
	}

	// Generate Excel report
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing Excel file: %v", err)
		}
	}()

	// Set headers
	headers := []string{"ID", "Asset Name", "Asset Type", "Institution", "Department", "Functional Area", "Manufacturer", "Model Number", "Serial Number", "Location", "Status", "Purchase Date", "Purchase Price", "Created At"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		if err := f.SetCellValue("Sheet1", cell, header); err != nil {
			log.Printf("Error setting header cell %s: %v", cell, err)
		}
	}

	// Add data
	for i, asset := range assets {
		row := i + 2
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), asset.ID); err != nil { log.Printf("SetCellValue A: %v", err) }
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), asset.AssetName); err != nil { log.Printf("SetCellValue B: %v", err) }
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), asset.AssetType); err != nil { log.Printf("SetCellValue C: %v", err) }
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), asset.InstitutionName); err != nil { log.Printf("SetCellValue D: %v", err) }
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), asset.Department); err != nil { log.Printf("SetCellValue E: %v", err) }
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("F%d", row), asset.FunctionalArea); err != nil { log.Printf("SetCellValue F: %v", err) }
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("G%d", row), asset.Manufacturer); err != nil { log.Printf("SetCellValue G: %v", err) }
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("H%d", row), asset.ModelNumber); err != nil { log.Printf("SetCellValue H: %v", err) }
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("I%d", row), asset.SerialNumber); err != nil { log.Printf("SetCellValue I: %v", err) }
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("J%d", row), asset.Location); err != nil { log.Printf("SetCellValue J: %v", err) }
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("K%d", row), asset.Status); err != nil { log.Printf("SetCellValue K: %v", err) }
		if asset.PurchaseDate != nil { if err := f.SetCellValue("Sheet1", fmt.Sprintf("L%d", row), asset.PurchaseDate.Format("2006-01-02")); err != nil { log.Printf("SetCellValue L: %v", err) } }
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("M%d", row), asset.PurchasePrice); err != nil { log.Printf("SetCellValue M: %v", err) }
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("N%d", row), asset.CreatedAt.Format("2006-01-02 15:04:05")); err != nil { log.Printf("SetCellValue N: %v", err) }
	}

	// Save file
	filename := fmt.Sprintf("asset_report_%s.xlsx", time.Now().Format("20060102_150405"))
	if err := f.SaveAs(filename); err != nil {
		log.Printf("Error saving Excel file: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Asset report generated successfully",
		Data: gin.H{
			"filename":   filename,
			"assetCount": len(assets),
		},
	})
}

// generateInvoiceHandler generates an invoice PDF
func generateInvoiceHandler(c *gin.Context) {
	var req InvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(190, 10, "INVOICE")
	pdf.Ln(15)

	// Customer information
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "Bill To:")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 6, req.CustomerName)
	pdf.Ln(6)
	pdf.Cell(190, 6, req.CustomerAddress)
	pdf.Ln(15)

	// Invoice details
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 8, fmt.Sprintf("Invoice Date: %s", time.Now().Format("January 2, 2006")))
	pdf.Ln(8)
	pdf.Cell(190, 8, fmt.Sprintf("Invoice Number: INV-%s", time.Now().Format("20060102")))
	pdf.Ln(15)

	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(80, 8, "Description")
	pdf.Cell(30, 8, "Quantity")
	pdf.Cell(40, 8, "Unit Price")
	pdf.Cell(40, 8, "Total")
	pdf.Ln(8)

	// Table content
	pdf.SetFont("Arial", "", 10)
	var totalAmount float64
	for _, item := range req.Items {
		itemTotal := float64(item.Quantity) * item.UnitPrice
		totalAmount += itemTotal

		pdf.Cell(80, 6, item.Description)
		pdf.Cell(30, 6, strconv.Itoa(item.Quantity))
		pdf.Cell(40, 6, formatCurrency(item.UnitPrice))
		pdf.Cell(40, 6, formatCurrency(itemTotal))
		pdf.Ln(6)
	}

	// Total
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(150, 8, "Total:")
	pdf.Cell(40, 8, formatCurrency(totalAmount))
	pdf.Ln(15)

	// Amount in words
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 6, fmt.Sprintf("Amount in words: %s pesos only", numberToWords(int(totalAmount))))
	pdf.Ln(20)

	// Terms and conditions
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 8, "Terms and Conditions:")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 8)
	pdf.Cell(190, 5, "1. Payment is due within 30 days of invoice date.")
	pdf.Ln(5)
	pdf.Cell(190, 5, "2. Late payments may incur additional charges.")
	pdf.Ln(5)
	pdf.Cell(190, 5, "3. All disputes will be resolved through proper channels.")

	// Save PDF
	filename := fmt.Sprintf("invoice_%s.pdf", time.Now().Format("20060102_150405"))
	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		log.Printf("Error saving PDF: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Invoice generated successfully",
		Data: gin.H{
			"filename":     filename,
			"totalAmount":  totalAmount,
			"itemCount":    len(req.Items),
			"customerName": req.CustomerName,
		},
	})
}

// downloadHandler serves file downloads
func downloadHandler(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Filename is required",
		})
		return
	}

	// Set headers for file download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	// Serve the file
	c.File(filename)
}

// fetchAssetsByInstitutionHandler fetches assets by institution for Excel reports
func fetchAssetsByInstitutionHandler(c *gin.Context) {
	var req InstitutionBarcodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Build query with institution filter
	query := `
		SELECT id, asset_name, asset_type, institution_name, department, functional_area, 
		manufacturer, model_number, serial_number, location, status, purchase_date, 
		purchase_price, logo, created_at, updated_at 
		FROM assets 
		WHERE institution_name = ? AND institution_name IS NOT NULL
		ORDER BY asset_name
	`

	rows, err := db.Query(query, req.Institution)
	if err != nil {
		log.Printf("Error fetching assets by institution: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to fetch assets: " + err.Error(),
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

		// Calculate market value (use purchase price as fallback)
		marketValue := 0.0
		if asset.PurchasePrice != nil {
			marketValue = *asset.PurchasePrice
		}

		assets = append(assets, gin.H{
			"id":              asset.ID,
			"assetName":       asset.AssetName,
			"assetType":       asset.AssetType,
			"institutionName": asset.InstitutionName,
			"department":      asset.Department,
			"functionalArea":  asset.FunctionalArea,
			"marketValue":     marketValue,
			"location":        asset.Location,
			"status":          asset.Status,
			"manufacturer":    asset.Manufacturer,
			"modelNumber":     asset.ModelNumber,
			"serialNumber":    asset.SerialNumber,
			"purchaseDate":    asset.PurchaseDate.Format("2006-01-02"),
			"purchasePrice":   asset.PurchasePrice,
			"createdAt":       asset.CreatedAt.Format("2006-01-02 15:04:05"),
			"updatedAt":       asset.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating assets: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Error processing assets",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    assets,
	})
} 