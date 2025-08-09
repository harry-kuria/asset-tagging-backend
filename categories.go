package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// getCategoriesHandler returns all categories for the current company
func getCategoriesHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)

	rows, err := db.Query(`
		SELECT id, company_id, name, description, color, is_active, created_at, updated_at
		FROM asset_categories 
		WHERE company_id = ? AND is_active = true
		ORDER BY name
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to fetch categories: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var categories []AssetCategory
	for rows.Next() {
		var cat AssetCategory
		err := rows.Scan(
			&cat.ID, &cat.CompanyID, &cat.Name, &cat.Description,
			&cat.Color, &cat.IsActive, &cat.CreatedAt, &cat.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   "Failed to scan category: " + err.Error(),
			})
			return
		}
		categories = append(categories, cat)
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    categories,
	})
}

// addCategoryHandler adds a new category for the current company
func addCategoryHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)

	var req AddCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Check if category name already exists for this company
	var existingID int
	err := db.QueryRow("SELECT id FROM asset_categories WHERE company_id = ? AND name = ?", companyID, req.Name).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Error:   "Category name already exists",
		})
		return
	}

	// Set default color if not provided
	if req.Color == "" {
		req.Color = "#007bff"
	}

	result, err := db.Exec(`
		INSERT INTO asset_categories (company_id, name, description, color)
		VALUES (?, ?, ?, ?)
	`, companyID, req.Name, req.Description, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to create category: " + err.Error(),
		})
		return
	}

	categoryID, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Category created successfully",
		Data: map[string]interface{}{
			"id": categoryID,
		},
	})
}

// updateCategoryHandler updates an existing category
func updateCategoryHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid category ID",
		})
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Check if category exists and belongs to this company
	var existingID int
	err = db.QueryRow("SELECT id FROM asset_categories WHERE id = ? AND company_id = ?", categoryID, companyID).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Error:   "Category not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database error: " + err.Error(),
		})
		return
	}

	// Build update query dynamically
	query := "UPDATE asset_categories SET "
	var params []interface{}
	var updates []string

	if req.Name != "" {
		updates = append(updates, "name = ?")
		params = append(params, req.Name)
	}
	if req.Description != "" {
		updates = append(updates, "description = ?")
		params = append(params, req.Description)
	}
	if req.Color != "" {
		updates = append(updates, "color = ?")
		params = append(params, req.Color)
	}
	if req.IsActive != nil {
		updates = append(updates, "is_active = ?")
		params = append(params, *req.IsActive)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "No fields to update",
		})
		return
	}

	query += updates[0]
	for i := 1; i < len(updates); i++ {
		query += ", " + updates[i]
	}
	query += ", updated_at = NOW() WHERE id = ? AND company_id = ?"
	params = append(params, categoryID, companyID)

	_, err = db.Exec(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to update category: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Category updated successfully",
	})
}

// deleteCategoryHandler deletes a category (soft delete by setting is_active = false)
func deleteCategoryHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid category ID",
		})
		return
	}

	// Check if category exists and belongs to this company
	var existingID int
	err = db.QueryRow("SELECT id FROM asset_categories WHERE id = ? AND company_id = ?", categoryID, companyID).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Error:   "Category not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database error: " + err.Error(),
		})
		return
	}

	// Check if category is being used by any assets
	var assetCount int
	err = db.QueryRow("SELECT COUNT(*) FROM assets WHERE category_id = ? AND company_id = ?", categoryID, companyID).Scan(&assetCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to check category usage: " + err.Error(),
		})
		return
	}

	if assetCount > 0 {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Error:   "Cannot delete category that is being used by assets",
		})
		return
	}

	// Soft delete by setting is_active = false
	_, err = db.Exec("UPDATE asset_categories SET is_active = false, updated_at = NOW() WHERE id = ? AND company_id = ?", categoryID, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to delete category: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Category deleted successfully",
	})
} 