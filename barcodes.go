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

	query := fmt.Sprintf("SELECT id, assetName, assetType, institutionName, department, functionalArea, manufacturer, modelNumber, serialNumber, location, status, purchaseDate, purchasePrice, logo, createdAt, updatedAt FROM assets WHERE id IN (%s)", placeholders)

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
			&asset.Logo, &asset.CreatedAt, &asset.UpdatedAt)
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
		barcodeFilename := fmt.Sprintf("barcode_%d.png", asset.ID)
		file, err := os.Create(barcodeFilename)
		if err != nil {
			log.Printf("Error creating barcode file: %v", err)
			continue
		}
		
		err = png.Encode(file, scaledCode)
		file.Close()
		if err != nil {
			log.Printf("Error saving barcode: %v", err)
			continue
		}

		// Add barcode to PDF
		pdf.Image(barcodeFilename, 10, float64(30+(i%2)*120), 80, 20, false, "", 0, "")
		
		// Add asset details
		pdf.SetXY(10, float64(55+(i%2)*120))
		pdf.Cell(80, 5, fmt.Sprintf("Asset: %s", asset.AssetName))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Type: %s", asset.AssetType))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Institution: %s", asset.InstitutionName))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Department: %s", asset.Department))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Location: %s", asset.Location))
		pdf.Ln(10)

		// Clean up temporary file
		os.Remove(barcodeFilename)
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
		SELECT id, assetName, assetType, institutionName, department, functionalArea, 
		manufacturer, modelNumber, serialNumber, location, status, purchaseDate, 
		purchasePrice, logo, createdAt, updatedAt 
		FROM assets WHERE institutionName = ?`, req.Institution)
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
			&asset.Logo, &asset.CreatedAt, &asset.UpdatedAt)
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
		barcodeFilename := fmt.Sprintf("barcode_%d.png", asset.ID)
		file, err := os.Create(barcodeFilename)
		if err != nil {
			log.Printf("Error creating barcode file: %v", err)
			continue
		}
		
		err = png.Encode(file, scaledCode)
		file.Close()
		if err != nil {
			log.Printf("Error saving barcode: %v", err)
			continue
		}

		// Add barcode to PDF
		pdf.Image(barcodeFilename, 10, float64(30+(i%2)*120), 80, 20, false, "", 0, "")
		
		// Add asset details
		pdf.SetXY(10, float64(55+(i%2)*120))
		pdf.Cell(80, 5, fmt.Sprintf("Asset: %s", asset.AssetName))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Type: %s", asset.AssetType))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Department: %s", asset.Department))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Location: %s", asset.Location))
		pdf.Ln(10)

		// Clean up temporary file
		os.Remove(barcodeFilename)
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

	// Get all assets for the institution and department
	rows, err := db.Query(`
		SELECT id, assetName, assetType, institutionName, department, functionalArea, 
		manufacturer, modelNumber, serialNumber, location, status, purchaseDate, 
		purchasePrice, logo, createdAt, updatedAt 
		FROM assets WHERE institutionName = ? AND department = ?`, req.Institution, req.Department)
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
			&asset.Logo, &asset.CreatedAt, &asset.UpdatedAt)
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
		barcodeFilename := fmt.Sprintf("barcode_%d.png", asset.ID)
		file, err := os.Create(barcodeFilename)
		if err != nil {
			log.Printf("Error creating barcode file: %v", err)
			continue
		}
		
		err = png.Encode(file, scaledCode)
		file.Close()
		if err != nil {
			log.Printf("Error saving barcode: %v", err)
			continue
		}

		// Add barcode to PDF
		pdf.Image(barcodeFilename, 10, float64(30+(i%2)*120), 80, 20, false, "", 0, "")
		
		// Add asset details
		pdf.SetXY(10, float64(55+(i%2)*120))
		pdf.Cell(80, 5, fmt.Sprintf("Asset: %s", asset.AssetName))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Type: %s", asset.AssetType))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Location: %s", asset.Location))
		pdf.Ln(5)
		pdf.Cell(80, 5, fmt.Sprintf("Status: %s", asset.Status))
		pdf.Ln(10)

		// Clean up temporary file
		os.Remove(barcodeFilename)
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
		},
	})
}

// generateBarcodeData creates the data string for barcode generation
func generateBarcodeData(asset Asset) string {
	// Create a formatted string with asset information
	data := fmt.Sprintf("ID:%d|Name:%s|Type:%s|Inst:%s|Dept:%s|Loc:%s",
		asset.ID,
		getShortName(asset.AssetName),
		getShortForm(asset.AssetType),
		getInstitutionInitials(asset.InstitutionName),
		getShortName(asset.Department),
		getShortName(asset.Location))
	
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

// fetchAssetsByInstitutionHandler fetches assets by institution
func fetchAssetsByInstitutionHandler(c *gin.Context) {
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
		SELECT id, assetName, assetType, institutionName, department, functionalArea, 
		manufacturer, modelNumber, serialNumber, location, status, purchaseDate, 
		purchasePrice, logo, createdAt, updatedAt 
		FROM assets WHERE institutionName = ?`, req.Institution)
	if err != nil {
		log.Printf("Error fetching assets by institution: %v", err)
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

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: gin.H{
			"assets":      assets,
			"institution": req.Institution,
			"count":       len(assets),
		},
	})
} 