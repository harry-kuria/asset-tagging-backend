package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// createCompanyHandler creates a new company with auto-incrementing ID
func createCompanyHandler(c *gin.Context) {
	var req RegisterCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Generate company code if not provided
	if req.CompanyCode == "" {
		req.CompanyCode = generateCompanyCode(req.CompanyName)
	}

	// Check if company code already exists
	var existingCompanyID int
	err := db.QueryRow("SELECT id FROM companies WHERE company_code = ?", req.CompanyCode).Scan(&existingCompanyID)
	if err == nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Error:   "Company code already exists",
		})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database error: " + err.Error(),
		})
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to start transaction: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// Create company
	result, err := tx.Exec(`
		INSERT INTO companies (company_name, company_code, email, subscription_plan, is_active, trial_ends_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, req.CompanyName, req.CompanyCode, req.Email, "trial", true, time.Now().AddDate(0, 0, 30), time.Now(), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to create company: " + err.Error(),
		})
		return
	}

	companyID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get company ID: " + err.Error(),
		})
		return
	}

	// Hash password for admin user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminUser.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to hash password: " + err.Error(),
		})
		return
	}

	// Create admin user
	userResult, err := tx.Exec(`
		INSERT INTO users (company_id, username, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, companyID, req.AdminUser.Username, req.AdminUser.Email, string(hashedPassword), 
	   req.AdminUser.FirstName, req.AdminUser.LastName, "admin", true, time.Now(), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to create admin user: " + err.Error(),
		})
		return
	}

	userID, err := userResult.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get user ID: " + err.Error(),
		})
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to commit transaction: " + err.Error(),
		})
		return
	}

	// Generate JWT token for the new admin user
	token, err := generateJWT(int(userID), int(companyID), req.AdminUser.Username, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to generate token: " + err.Error(),
		})
		return
	}

	// Get the created company and user
	var company Company
	err = db.QueryRow(`
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
			Error:   "Failed to fetch company details: " + err.Error(),
		})
		return
	}

	var user User
	err = db.QueryRow(`
		SELECT id, company_id, username, email, first_name, last_name, role, is_active, created_at, updated_at
		FROM users WHERE id = ?
	`, userID).Scan(
		&user.ID, &user.CompanyID, &user.Username, &user.Email,
		&user.FirstName, &user.LastName, &user.Role, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to fetch user details: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Company created successfully",
		Data: LoginResponse{
			Token:     token,
			User:      user,
			Company:   company,
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			Roles:     []string{"admin"},
		},
	})
}

// generateCompanyCode generates a unique company code from company name
func generateCompanyCode(companyName string) string {
	// Remove special characters and convert to uppercase
	code := strings.ToUpper(strings.ReplaceAll(companyName, " ", ""))
	code = strings.ReplaceAll(code, "-", "")
	code = strings.ReplaceAll(code, "_", "")
	code = strings.ReplaceAll(code, ".", "")
	code = strings.ReplaceAll(code, ",", "")
	
	// Take first 6 characters
	if len(code) > 6 {
		code = code[:6]
	}
	
	// Add timestamp suffix to ensure uniqueness
	timestamp := fmt.Sprintf("%d", time.Now().Unix() % 10000)
	return code + timestamp
}

// getCompanyHandler returns company details
func getCompanyHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)

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
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Error:   "Company not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    company,
	})
}

// updateCompanyHandler updates company details
func updateCompanyHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)

	var req struct {
		CompanyName string `json:"company_name"`
		Email       string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request data: " + err.Error(),
		})
		return
	}

	_, err := db.Exec(`
		UPDATE companies 
		SET company_name = ?, email = ?, updated_at = ?
		WHERE id = ?
	`, req.CompanyName, req.Email, time.Now(), companyID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to update company: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Company updated successfully",
	})
}

// listCompaniesHandler returns all companies (admin only)
func listCompaniesHandler(c *gin.Context) {
	// Check if user is admin
	user := getCurrentUser(c)
	if user == nil || user.Role != "admin" {
		c.JSON(http.StatusForbidden, APIResponse{
			Success: false,
			Error:   "Access denied",
		})
		return
	}

	rows, err := db.Query(`
		SELECT id, company_name, company_code, email, subscription_plan, is_active, trial_ends_at, created_at, updated_at
		FROM companies
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database error: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var company Company
		err := rows.Scan(
			&company.ID, &company.CompanyName, &company.CompanyCode, &company.Email,
			&company.SubscriptionPlan, &company.IsActive, &company.TrialEndsAt,
			&company.CreatedAt, &company.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning company: %v", err)
			continue
		}
		companies = append(companies, company)
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    companies,
	})
} 