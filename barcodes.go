package main

import (
	"fmt"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

// safeString safely converts a nullable string to a regular string
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// generateBarcodesHandler generates barcodes for specific assets
func generateBarcodesHandler(c *gin.Context) {
	var req BarcodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	// Get asset details for the provided IDs
	placeholders := strings.Repeat("?,", len(req.AssetIDs))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf("SELECT id, asset_name, asset_type, institution_name, department, functional_area, manufacturer, model_number, serial_number, location, status, purchase_date, purchase_price, created_at, updated_at FROM assets WHERE id IN (%s)", placeholders)

	args := make([]interface{}, len(req.AssetIDs))
	for i, id := range req.AssetIDs {
		args[i] = id
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error fetching assets for barcode generation: %v", err)
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
			&asset.CreatedAt, &asset.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning asset: %v", err)
			continue
		}
		assets = append(assets, asset)
	}

	// Generate PDF with barcodes
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Asset Barcodes")
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 10)

	for i, asset := range assets {
		if i > 0 && i%2 == 0 {
			pdf.AddPage()
		}

		// Generate barcode data
		barcodeData := generateBarcodeData(asset)

		// Create barcode
		code, err := code128.Encode(barcodeData)
		if err != nil {
			log.Printf("Error creating barcode: %v", err)
			continue
		}

		// Scale barcode
		scaledCode, err := barcode.Scale(code, 200, 50)
		if err != nil {
			log.Printf("Error scaling barcode: %v", err)
			continue
		}

		// Save barcode as image
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("barcode_%d_*.png", asset.ID))
		if err != nil {
			log.Printf("Error creating barcode temp file: %v", err)
			continue
		}
		if err := png.Encode(tmpFile, scaledCode); err != nil {
			_ = tmpFile.Close()
			log.Printf("Error encoding barcode: %v", err)
			continue
		}
		if err := tmpFile.Close(); err != nil {
			log.Printf("Error closing barcode file: %v", err)
		}

		// Add barcode to PDF
		pdf.Image(tmpFile.Name(), 10, float64(30+(i%2)*120), 80, 20, false, "", 0, "")
		
		// Add asset details
		pdf.SetXY(10, float64(55+(i%2)*120))
		pdf.Cell(80, 5, fmt.Sprintf("Asset: %s", asset.AssetName))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Type: %s", safeString(asset.AssetType)))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Institution: %s", safeString(asset.InstitutionName)))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Department: %s", safeString(asset.Department)))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Location: %s", safeString(asset.Location)))
		pdf.Ln(10)

		// Clean up temporary file
		if err := os.Remove(tmpFile.Name()); err != nil {
			log.Printf("Error removing temp barcode file: %v", err)
		}
	}

	// Save PDF
	pdfFilename := "asset_barcodes.pdf"
	err = pdf.OutputFileAndClose(pdfFilename)
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
		Message: "Barcodes generated successfully",
		Data: gin.H{
			"filename": pdfFilename,
			"assetCount": len(assets),
		},
	})
}

// generateBarcodesByInstitutionHandler generates barcodes for all assets in an institution
func generateBarcodesByInstitutionHandler(c *gin.Context) {
	var req InstitutionBarcodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	// Get current company ID
	companyID := getCurrentCompanyID(c)

	// Get all assets for the institution within the current company
	rows, err := db.Query(`
		SELECT id, company_id, asset_name, asset_type, institution_name, department, functional_area, 
		manufacturer, model_number, serial_number, location, status, purchase_date, 
		purchase_price, created_at, updated_at 
		FROM assets WHERE institution_name = ? AND company_id = ?`, req.Institution, companyID)
	if err != nil {
		log.Printf("Error fetching assets by institution: %v", err)
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
			&asset.ID, &asset.CompanyID, &asset.AssetName, &asset.AssetType, &asset.InstitutionName, &asset.Department,
			&asset.FunctionalArea, &asset.Manufacturer, &asset.ModelNumber, &asset.SerialNumber,
			&asset.Location, &asset.Status, &asset.PurchaseDate, &asset.PurchasePrice,
			&asset.CreatedAt, &asset.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning asset: %v", err)
			continue
		}
		assets = append(assets, asset)
	}

	// Debug logging
	log.Printf("Found %d assets for institution '%s' in company %d", len(assets), req.Institution, companyID)
	
	// Enhanced debugging to understand the data
	log.Printf("=== DEBUG: Barcode Generation for Institution '%s' ===", req.Institution)
	log.Printf("Company ID: %d", companyID)
	log.Printf("Total assets found: %d", len(assets))
	
	// Check total assets in company
	var totalCompanyAssets int
	err = db.QueryRow("SELECT COUNT(*) FROM assets WHERE company_id = ?", companyID).Scan(&totalCompanyAssets)
	if err == nil {
		log.Printf("Total assets in company %d: %d", companyID, totalCompanyAssets)
	}
	
	// Check total assets for this institution across all companies
	var totalInstitutionAssets int
	err = db.QueryRow("SELECT COUNT(*) FROM assets WHERE institution_name = ?", req.Institution).Scan(&totalInstitutionAssets)
	if err == nil {
		log.Printf("Total assets for institution '%s' across all companies: %d", req.Institution, totalInstitutionAssets)
	}
	
	if len(assets) == 0 {
		// Log all assets for this company to see what's available
		allRows, err := db.Query("SELECT institution_name, department FROM assets WHERE company_id = ?", companyID)
		if err == nil {
			defer allRows.Close()
			log.Printf("Available institution/department combinations for company %d:", companyID)
			for allRows.Next() {
				var inst, dept *string
				if err := allRows.Scan(&inst, &dept); err == nil {
					instStr := "NULL"
					deptStr := "NULL"
					if inst != nil {
						instStr = *inst
					}
					if dept != nil {
						deptStr = *dept
					}
					log.Printf("  Institution: '%s', Department: '%s'", instStr, deptStr)
				}
			}
		}
		
		// Also check what institutions exist in the database
		instRows, err := db.Query("SELECT DISTINCT institution_name FROM assets WHERE company_id = ? AND institution_name IS NOT NULL", companyID)
		if err == nil {
			defer instRows.Close()
			log.Printf("Available institutions for company %d:", companyID)
			for instRows.Next() {
				var inst *string
				if err := instRows.Scan(&inst); err == nil {
					instStr := "NULL"
					if inst != nil {
						instStr = *inst
					}
					log.Printf("  Institution: '%s'", instStr)
				}
			}
		}
	} else {
		// Log sample of found assets
		log.Printf("Sample of found assets:")
		for i, asset := range assets {
			if i < 5 { // Show first 5 assets
				log.Printf("  Asset %d: ID=%d, Name='%s', Institution='%s', Department='%s'", 
					i+1, asset.ID, asset.AssetName, safeString(asset.InstitutionName), safeString(asset.Department))
			}
		}
		if len(assets) > 5 {
			log.Printf("  ... and %d more assets", len(assets)-5)
		}
	}
	
	log.Printf("=== END DEBUG ===")

	// Generate PDF with barcodes
	pdf := gofpdf.New("P", "mm", "A4", "")
	
	// Calculate how many barcodes per page (2 columns, 2 rows = 4 per page)
	barcodesPerPage := 4
	totalPages := (len(assets) + barcodesPerPage - 1) / barcodesPerPage
	
	log.Printf("Generating PDF with %d assets across %d pages (%d barcodes per page)", len(assets), totalPages, barcodesPerPage)

	for pageNum := 0; pageNum < totalPages; pageNum++ {
		pdf.AddPage()
		
		// Set font for header
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(190, 10, fmt.Sprintf("Asset Barcodes - %s (Page %d of %d)", req.Institution, pageNum+1, totalPages))
		pdf.Ln(15)

		// Set font for details
		pdf.SetFont("Arial", "", 10)
		
		// Calculate start and end indices for this page
		startIdx := pageNum * barcodesPerPage
		endIdx := startIdx + barcodesPerPage
		if endIdx > len(assets) {
			endIdx = len(assets)
		}
		
		// Process assets for this page
		for i := startIdx; i < endIdx; i++ {
			asset := assets[i]
			position := i - startIdx // Position on this page (0-3)
			
			// Calculate position on page (2x2 grid)
			row := position / 2
			col := position % 2
			
			xPos := float64(10 + col*95) // 95mm spacing between columns
			yPos := float64(30 + row*120) // 120mm spacing between rows

			// Generate barcode data
			barcodeData := generateBarcodeData(asset)

			// Create barcode
			code, err := code128.Encode(barcodeData)
			if err != nil {
				log.Printf("Error creating barcode for asset %d: %v", asset.ID, err)
				continue
			}

			// Scale barcode
			scaledCode, err := barcode.Scale(code, 200, 50)
			if err != nil {
				log.Printf("Error scaling barcode for asset %d: %v", asset.ID, err)
				continue
			}

			// Save barcode as image
			tmpFile, err := os.CreateTemp("", fmt.Sprintf("barcode_%d_*.png", asset.ID))
			if err != nil {
				log.Printf("Error creating barcode temp file for asset %d: %v", asset.ID, err)
				continue
			}
			if err := png.Encode(tmpFile, scaledCode); err != nil {
				_ = tmpFile.Close()
				log.Printf("Error encoding barcode for asset %d: %v", asset.ID, err)
				continue
			}
			if err := tmpFile.Close(); err != nil {
				log.Printf("Error closing barcode file for asset %d: %v", asset.ID, err)
			}

			// Add barcode to PDF
			pdf.Image(tmpFile.Name(), xPos, yPos, 80, 20, false, "", 0, "")
			
			// Add asset details below barcode
			pdf.SetXY(xPos, yPos+25.0)
			pdf.Cell(80, 5, fmt.Sprintf("Asset: %s", asset.AssetName))
			pdf.Ln(5)
			pdf.Cell(80, 5, fmt.Sprintf("Type: %s", safeString(asset.AssetType)))
			pdf.Ln(5)
			pdf.Cell(80, 5, fmt.Sprintf("Department: %s", safeString(asset.Department)))
			pdf.Ln(5)
			pdf.Cell(80, 5, fmt.Sprintf("Location: %s", safeString(asset.Location)))
			pdf.Ln(10)

			// Clean up temporary file immediately
			if err := os.Remove(tmpFile.Name()); err != nil {
				log.Printf("Error removing temp barcode file for asset %d: %v", asset.ID, err)
			}
		}
		
		// Add progress indicator
		log.Printf("Generated page %d/%d with %d barcodes", pageNum+1, totalPages, endIdx-startIdx)
	}

	// Save PDF
	pdfFilename := fmt.Sprintf("barcodes_%s.pdf", strings.ReplaceAll(req.Institution, " ", "_"))
	err = pdf.OutputFileAndClose(pdfFilename)
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
		Message: fmt.Sprintf("Barcodes generated successfully for %d assets", len(assets)),
		Data: gin.H{
			"filename":   pdfFilename,
			"assetCount": len(assets),
			"institution": req.Institution,
			"barcodeTags": generateBarcodeTags(assets),
			"assetDetails": assets,
			"totalPages": totalPages,
			"barcodesPerPage": barcodesPerPage,
			"generationTime": "Completed", // Could add actual timing if needed
		},
	})
}

// generateBarcodesByInstitutionAndDepartmentHandler generates barcodes for assets in a specific institution and department
func generateBarcodesByInstitutionAndDepartmentHandler(c *gin.Context) {
	var req InstitutionDepartmentBarcodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	// Debug logging
	log.Printf("Received barcode generation request - Institution: '%s', Department: '%s'", req.Institution, req.Department)

	// Get current company ID
	companyID := getCurrentCompanyID(c)

	// Get all assets for the institution and department within the current company
	rows, err := db.Query(`
		SELECT id, company_id, asset_name, asset_type, institution_name, department, functional_area, 
		manufacturer, model_number, serial_number, location, status, purchase_date, 
		purchase_price, created_at, updated_at 
		FROM assets WHERE institution_name = ? AND department = ? AND company_id = ?`, req.Institution, req.Department, companyID)
	if err != nil {
		log.Printf("Error fetching assets by institution and department: %v", err)
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
			&asset.ID, &asset.CompanyID, &asset.AssetName, &asset.AssetType, &asset.InstitutionName, &asset.Department,
			&asset.FunctionalArea, &asset.Manufacturer, &asset.ModelNumber, &asset.SerialNumber,
			&asset.Location, &asset.Status, &asset.PurchaseDate, &asset.PurchasePrice,
			&asset.CreatedAt, &asset.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning asset: %v", err)
			continue
		}
		assets = append(assets, asset)
	}

	// Debug logging
	log.Printf("Found %d assets for institution '%s' and department '%s' in company %d", len(assets), req.Institution, req.Department, companyID)
	if len(assets) == 0 {
		// Log all assets for this company to see what's available
		allRows, err := db.Query("SELECT institution_name, department FROM assets WHERE company_id = ?", companyID)
		if err == nil {
			defer allRows.Close()
			log.Printf("Available institution/department combinations for company %d:", companyID)
			for allRows.Next() {
				var inst, dept *string
				if err := allRows.Scan(&inst, &dept); err == nil {
					instStr := "NULL"
					deptStr := "NULL"
					if inst != nil {
						instStr = *inst
					}
					if dept != nil {
						deptStr = *dept
					}
					log.Printf("  Institution: '%s', Department: '%s'", instStr, deptStr)
				}
			}
		}
	}

	// Generate PDF with barcodes
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, fmt.Sprintf("Asset Barcodes - %s - %s", req.Institution, req.Department))
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 10)

	for i, asset := range assets {
		if i > 0 && i%2 == 0 {
			pdf.AddPage()
		}

		// Generate barcode data
		barcodeData := generateBarcodeData(asset)

		// Create barcode
		code, err := code128.Encode(barcodeData)
		if err != nil {
			log.Printf("Error creating barcode: %v", err)
			continue
		}

		// Scale barcode
		scaledCode, err := barcode.Scale(code, 200, 50)
		if err != nil {
			log.Printf("Error scaling barcode: %v", err)
			continue
		}

		// Save barcode as image
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("barcode_%d_*.png", asset.ID))
		if err != nil {
			log.Printf("Error creating barcode temp file: %v", err)
			continue
		}
		if err := png.Encode(tmpFile, scaledCode); err != nil {
			_ = tmpFile.Close()
			log.Printf("Error encoding barcode: %v", err)
			continue
		}
		if err := tmpFile.Close(); err != nil {
			log.Printf("Error closing barcode file: %v", err)
		}

		// Add barcode to PDF
		pdf.Image(tmpFile.Name(), 10, float64(30+(i%2)*120), 80, 20, false, "", 0, "")
		
		// Add asset details
		pdf.SetXY(10, float64(55+(i%2)*120))
		pdf.Cell(80, 5, fmt.Sprintf("Asset: %s", asset.AssetName))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Type: %s", safeString(asset.AssetType)))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Location: %s", safeString(asset.Location)))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Status: %s", asset.Status))
		pdf.Ln(10)

		// Clean up temporary file
		if err := os.Remove(tmpFile.Name()); err != nil {
			log.Printf("Error removing temp barcode file: %v", err)
		}
	}

	// Save PDF
	pdfFilename := fmt.Sprintf("barcodes_%s_%s.pdf", 
		strings.ReplaceAll(req.Institution, " ", "_"),
		strings.ReplaceAll(req.Department, " ", "_"))
	err = pdf.OutputFileAndClose(pdfFilename)
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
		Message: "Barcodes generated successfully",
		Data: gin.H{
			"filename":   pdfFilename,
			"assetCount": len(assets),
			"institution": req.Institution,
			"department":  req.Department,
			"barcodeTags": generateBarcodeTags(assets),
			"assetDetails": assets,
		},
	})
}

// generateBarcodeTags creates barcode tag data for frontend display
func generateBarcodeTags(assets []Asset) []gin.H {
	var barcodeTags []gin.H
	for _, asset := range assets {
		barcodeData := generateBarcodeData(asset)
		barcodeTags = append(barcodeTags, gin.H{
			"formattedString": barcodeData,
			"assetDetails": gin.H{
				"id":              asset.ID,
				"assetName":       asset.AssetName,
				"assetType":       safeString(asset.AssetType),
				"institutionName": safeString(asset.InstitutionName),
				"department":      safeString(asset.Department),
				"functionalArea":  safeString(asset.FunctionalArea),
				"manufacturer":    safeString(asset.Manufacturer),
				"modelNumber":     safeString(asset.ModelNumber),
				"serialNumber":    safeString(asset.SerialNumber),
				"location":        safeString(asset.Location),
				"status":          asset.Status,
				"purchaseDate":    asset.PurchaseDate,
				"purchasePrice":   asset.PurchasePrice,
			},
		})
	}
	return barcodeTags
}

// generateBarcodeData creates the data string for barcode generation
func generateBarcodeData(asset Asset) string {
	// Handle nullable fields
	assetType := ""
	if asset.AssetType != nil {
		assetType = *asset.AssetType
	}
	
	institutionName := ""
	if asset.InstitutionName != nil {
		institutionName = *asset.InstitutionName
	}
	
	department := ""
	if asset.Department != nil {
		department = *asset.Department
	}
	
	location := ""
	if asset.Location != nil {
		location = *asset.Location
	}

	// Create a formatted string with asset information
	data := fmt.Sprintf("ID:%d|Name:%s|Type:%s|Inst:%s|Dept:%s|Loc:%s",
		asset.ID,
		getShortName(asset.AssetName),
		getShortForm(assetType),
		getInstitutionInitials(institutionName),
		getShortName(department),
		getShortName(location))
	
	return data
}

// getShortName truncates a string to a reasonable length for barcode
func getShortName(fullText string) string {
	if len(fullText) <= 15 {
		return fullText
	}
	return fullText[:15]
}

// getShortForm gets a short form of asset type
func getShortForm(assetType string) string {
	shortForms := map[string]string{
		"Land":           "LND",
		"Building":       "BLD",
		"Equipment":      "EQP",
		"Vehicle":        "VEH",
		"Furniture":      "FRN",
		"Intangible":     "INT",
		"Biological":     "BIO",
		"Computer":       "COM",
		"Software":       "SW",
		"Network":        "NET",
		"Communication":  "COM",
		"Security":       "SEC",
		"Medical":        "MED",
		"Laboratory":     "LAB",
		"Office":         "OFF",
		"Kitchen":        "KIT",
		"Cleaning":       "CLN",
		"Maintenance":    "MNT",
		"Transportation": "TRN",
		"Recreation":     "REC",
		"Storage":        "STO",
		"Other":          "OTH",
	}

	if shortForm, exists := shortForms[assetType]; exists {
		return shortForm
	}
	return assetType[:3]
}

// getInstitutionInitials gets initials from institution name
func getInstitutionInitials(institutionName string) string {
	words := strings.Fields(institutionName)
	if len(words) == 0 {
		return "UNK"
	}
	
	var initials string
	for _, word := range words {
		if len(word) > 0 {
			initials += string(word[0])
		}
	}
	
	if len(initials) > 5 {
		return initials[:5]
	}
	return initials
} 

// generateBarcodesForAllInstitutionsHandler generates barcodes for all assets across all institutions in the company
func generateBarcodesForAllInstitutionsHandler(c *gin.Context) {
	// Get current company ID
	companyID := getCurrentCompanyID(c)

	// Get all assets for the company
	rows, err := db.Query(`
		SELECT id, company_id, asset_name, asset_type, institution_name, department, functional_area, 
		manufacturer, model_number, serial_number, location, status, purchase_date, 
		purchase_price, created_at, updated_at 
		FROM assets WHERE company_id = ? ORDER BY institution_name, department`, companyID)
	if err != nil {
		log.Printf("Error fetching all assets by company: %v", err)
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
			&asset.ID, &asset.CompanyID, &asset.AssetName, &asset.AssetType, &asset.InstitutionName, &asset.Department,
			&asset.FunctionalArea, &asset.Manufacturer, &asset.ModelNumber, &asset.SerialNumber,
			&asset.Location, &asset.Status, &asset.PurchaseDate, &asset.PurchasePrice,
			&asset.CreatedAt, &asset.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning asset: %v", err)
			continue
		}
		assets = append(assets, asset)
	}

	// Enhanced debugging for heavy load testing
	log.Printf("=== HEAVY LOAD TEST: Generating barcodes for ALL institutions ===")
	log.Printf("Company ID: %d", companyID)
	log.Printf("Total assets found: %d", len(assets))
	
	// Group assets by institution for better organization
	institutionAssets := make(map[string][]Asset)
	for _, asset := range assets {
		institution := safeString(asset.InstitutionName)
		if institution == "" {
			institution = "Unknown Institution"
		}
		institutionAssets[institution] = append(institutionAssets[institution], asset)
	}
	
	log.Printf("Assets grouped by institution:")
	for institution, instAssets := range institutionAssets {
		log.Printf("  %s: %d assets", institution, len(instAssets))
	}
	
	// Check total assets in company
	var totalCompanyAssets int
	err = db.QueryRow("SELECT COUNT(*) FROM assets WHERE company_id = ?", companyID).Scan(&totalCompanyAssets)
	if err == nil {
		log.Printf("Total assets in company %d: %d", companyID, totalCompanyAssets)
	}
	
	// Get unique institutions
	var uniqueInstitutions []string
	instRows, err := db.Query("SELECT DISTINCT institution_name FROM assets WHERE company_id = ? AND institution_name IS NOT NULL ORDER BY institution_name", companyID)
	if err == nil {
		defer instRows.Close()
		for instRows.Next() {
			var inst *string
			if err := instRows.Scan(&inst); err == nil {
				if inst != nil {
					uniqueInstitutions = append(uniqueInstitutions, *inst)
				}
			}
		}
	}
	log.Printf("Unique institutions found: %d", len(uniqueInstitutions))
	for _, inst := range uniqueInstitutions {
		log.Printf("  - %s", inst)
	}
	
	log.Printf("=== END HEAVY LOAD TEST DEBUG ===")

	// Generate PDF with barcodes organized by institution
	pdf := gofpdf.New("P", "mm", "A4", "")
	
	// Calculate pagination
	barcodesPerPage := 4
	totalPages := (len(assets) + barcodesPerPage - 1) / barcodesPerPage
	
	log.Printf("Generating comprehensive PDF with %d assets across %d pages (%d barcodes per page)", len(assets), totalPages, barcodesPerPage)

	currentPage := 0
	assetIndex := 0

	for _, institution := range uniqueInstitutions {
		instAssets := institutionAssets[institution]
		if len(instAssets) == 0 {
			continue
		}

		// Add institution header page
		pdf.AddPage()
		currentPage++
		
		// Set font for institution header
		pdf.SetFont("Arial", "B", 18)
		pdf.Cell(190, 15, fmt.Sprintf("Institution: %s", institution))
		pdf.Ln(20)
		
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 10, fmt.Sprintf("Total Assets: %d", len(instAssets)))
		pdf.Ln(15)
		
		// Add institution summary
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(190, 8, "Asset Summary:")
		pdf.Ln(10)
		
		// Group by department
		deptAssets := make(map[string][]Asset)
		for _, asset := range instAssets {
			dept := safeString(asset.Department)
			if dept == "" {
				dept = "Unknown Department"
			}
			deptAssets[dept] = append(deptAssets[dept], asset)
		}
		
		for dept, deptAssetList := range deptAssets {
			pdf.Cell(190, 6, fmt.Sprintf("  %s: %d assets", dept, len(deptAssetList)))
			pdf.Ln(6)
		}
		
		// Generate barcodes for this institution
		for _, asset := range instAssets {
			// Check if we need a new page for barcodes
			if assetIndex > 0 && assetIndex%barcodesPerPage == 0 {
				pdf.AddPage()
				currentPage++
				
				// Add page header
				pdf.SetFont("Arial", "B", 14)
				pdf.Cell(190, 10, fmt.Sprintf("Asset Barcodes - %s (Page %d)", institution, currentPage))
				pdf.Ln(15)
			}
			
			// If this is the first barcode on a new page, add header
			if assetIndex%barcodesPerPage == 0 {
				pdf.SetFont("Arial", "B", 12)
				pdf.Cell(190, 8, fmt.Sprintf("Institution: %s", institution))
				pdf.Ln(10)
				pdf.SetFont("Arial", "", 10)
			}

			position := assetIndex % barcodesPerPage
			
			// Calculate position on page (2x2 grid)
			row := position / 2
			col := position % 2
			
			xPos := float64(10 + col*95) // 95mm spacing between columns
			yPos := float64(40 + row*120) // 40mm offset for header, 120mm spacing between rows

			// Generate barcode data
			barcodeData := generateBarcodeData(asset)

			// Create barcode
			code, err := code128.Encode(barcodeData)
			if err != nil {
				log.Printf("Error creating barcode for asset %d: %v", asset.ID, err)
				assetIndex++
				continue
			}

			// Scale barcode
			scaledCode, err := barcode.Scale(code, 200, 50)
			if err != nil {
				log.Printf("Error scaling barcode for asset %d: %v", asset.ID, err)
				assetIndex++
				continue
			}

			// Save barcode as image
			tmpFile, err := os.CreateTemp("", fmt.Sprintf("barcode_%d_*.png", asset.ID))
			if err != nil {
				log.Printf("Error creating barcode temp file for asset %d: %v", asset.ID, err)
				assetIndex++
				continue
			}
			if err := png.Encode(tmpFile, scaledCode); err != nil {
				_ = tmpFile.Close()
				log.Printf("Error encoding barcode for asset %d: %v", asset.ID, err)
				assetIndex++
				continue
			}
			if err := tmpFile.Close(); err != nil {
				log.Printf("Error closing barcode file for asset %d: %v", asset.ID, err)
			}

			// Add barcode to PDF
			pdf.Image(tmpFile.Name(), xPos, yPos, 80, 20, false, "", 0, "")
			
			// Add asset details below barcode
			pdf.SetXY(xPos, yPos+25.0)
			pdf.Cell(80, 5, fmt.Sprintf("Asset: %s", asset.AssetName))
			pdf.Ln(5)
			pdf.Cell(80, 5, fmt.Sprintf("Type: %s", safeString(asset.AssetType)))
			pdf.Ln(5)
			pdf.Cell(80, 5, fmt.Sprintf("Department: %s", safeString(asset.Department)))
			pdf.Ln(5)
			pdf.Cell(80, 5, fmt.Sprintf("Location: %s", safeString(asset.Location)))
			pdf.Ln(10)

			// Clean up temporary file immediately
			if err := os.Remove(tmpFile.Name()); err != nil {
				log.Printf("Error removing temp barcode file for asset %d: %v", asset.ID, err)
			}
			
			assetIndex++
		}
		
		// Add progress indicator
		log.Printf("Generated barcodes for institution '%s' with %d assets", institution, len(instAssets))
	}
	
	log.Printf("Completed heavy load test: Generated %d total barcodes across %d institutions", len(assets), len(uniqueInstitutions))

	// Save PDF
	pdfFilename := fmt.Sprintf("all_institutions_barcodes_company_%d.pdf", companyID)
	err = pdf.OutputFileAndClose(pdfFilename)
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
		Message: fmt.Sprintf("Heavy load test completed: Generated barcodes for %d assets across %d institutions", len(assets), len(uniqueInstitutions)),
		Data: gin.H{
			"filename":   pdfFilename,
			"assetCount": len(assets),
			"institutionCount": len(uniqueInstitutions),
			"institutions": uniqueInstitutions,
			"barcodeTags": generateBarcodeTags(assets),
			"assetDetails": assets,
			"totalPages": currentPage,
			"barcodesPerPage": barcodesPerPage,
			"generationTime": "Heavy Load Test Completed",
		},
	})
} 