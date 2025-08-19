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

	// Get all assets for the institution
	rows, err := db.Query(`
		SELECT id, asset_name, asset_type, institution_name, department, functional_area, 
		manufacturer, model_number, serial_number, location, status, purchase_date, 
		purchase_price, created_at, updated_at 
		FROM assets WHERE institution_name = ?`, req.Institution)
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
	pdf.Cell(190, 10, fmt.Sprintf("Asset Barcodes - %s", req.Institution))
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
		Message: "Barcodes generated successfully",
		Data: gin.H{
			"filename":   pdfFilename,
			"assetCount": len(assets),
			"institution": req.Institution,
			"barcodeTags": generateBarcodeTags(assets),
			"assetDetails": assets,
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

	// Get all assets for the institution and department
	rows, err := db.Query(`
		SELECT id, asset_name, asset_type, institution_name, department, functional_area, 
		manufacturer, model_number, serial_number, location, status, purchase_date, 
		purchase_price, created_at, updated_at 
		FROM assets WHERE institution_name = ? AND department = ?`, req.Institution, req.Department)
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

	// Debug logging
	log.Printf("Found %d assets for institution '%s' and department '%s'", len(assets), req.Institution, req.Department)
	if len(assets) == 0 {
		// Log all assets to see what's available
		allRows, err := db.Query("SELECT institution_name, department FROM assets")
		if err == nil {
			defer allRows.Close()
			log.Printf("Available institution/department combinations:")
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