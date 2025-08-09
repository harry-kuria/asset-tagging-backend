package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims represents JWT claims with company support
type Claims struct {
	UserID    int    `json:"user_id"`
	CompanyID int    `json:"company_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

// loginHandler handles user login with company code
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request data: " + err.Error(),
		})
		return
	}

	// First, find the company by company code
	var company Company
	err := db.QueryRow(`
		SELECT id, company_name, company_code, email, subscription_plan, is_active, trial_ends_at 
		FROM companies 
		WHERE company_code = ? AND is_active = true
	`, req.CompanyCode).Scan(
		&company.ID, &company.CompanyName, &company.CompanyCode, 
		&company.Email, &company.SubscriptionPlan, &company.IsActive, &company.TrialEndsAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Error:   "Invalid company code",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database error: " + err.Error(),
		})
		return
	}

	// Check if company trial has expired
	if company.TrialEndsAt != nil && time.Now().After(*company.TrialEndsAt) {
		c.JSON(http.StatusForbidden, APIResponse{
			Success: false,
			Error:   "Company trial has expired. Please contact support.",
		})
		return
	}

	// Find user by username and company_id
	var user User
	err = db.QueryRow(`
		SELECT id, company_id, username, email, password_hash, first_name, last_name, role, is_active, last_login
		FROM users 
		WHERE username = ? AND company_id = ? AND is_active = true
	`, req.Username, company.ID).Scan(
		&user.ID, &user.CompanyID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Role, &user.IsActive, &user.LastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Error:   "Invalid username or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database error: " + err.Error(),
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Error:   "Invalid username or password",
		})
		return
	}

	// Update last login
	_, err = db.Exec("UPDATE users SET last_login = NOW() WHERE id = ?", user.ID)
	if err != nil {
		// Log error but don't fail login
		log.Printf("Failed to update last login: %v", err)
	}

	// Generate JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := Claims{
		UserID:    user.ID,
		CompanyID: user.CompanyID,
		Username:  user.Username,
		Role:      user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "your-secret-key"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	// Return login response
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Login successful",
		Data: LoginResponse{
			Token:     tokenString,
			User:      user,
			Company:   company,
			ExpiresAt: expirationTime.Unix(),
		},
	})
}

// registerCompanyHandler handles company registration with admin user
func registerCompanyHandler(c *gin.Context) {
	var req RegisterCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback()

	// Check if company code already exists
	var existingID int
	err = tx.QueryRow("SELECT id FROM companies WHERE company_code = ?", req.CompanyCode).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Error:   "Company code already exists",
		})
		return
	}

	// Check if company email already exists
	err = tx.QueryRow("SELECT id FROM companies WHERE email = ?", req.Email).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Error:   "Company email already exists",
		})
		return
	}

	// Create company
	var companyID int64
	result, err := tx.Exec(`
		INSERT INTO companies (company_name, company_code, email, phone, address, industry, trial_ends_at)
		VALUES (?, ?, ?, ?, ?, ?, DATE_ADD(NOW(), INTERVAL 30 DAY))
	`, req.CompanyName, req.CompanyCode, req.Email, req.Phone, req.Address, req.Industry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to create company: " + err.Error(),
		})
		return
	}

	companyID, err = result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get company ID",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminUser.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to hash password",
		})
		return
	}

	// Create admin user
	var userID int64
	result, err = tx.Exec(`
		INSERT INTO users (company_id, username, email, password_hash, first_name, last_name, role)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, companyID, req.AdminUser.Username, req.AdminUser.Email, string(hashedPassword),
		req.AdminUser.FirstName, req.AdminUser.LastName, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to create admin user: " + err.Error(),
		})
		return
	}

	userID, err = result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get user ID",
		})
		return
	}

	// Create default user roles for admin
	_, err = tx.Exec(`
		INSERT INTO user_roles (user_id, company_id, role) VALUES 
		(?, ?, 'userManagement'),
		(?, ?, 'assetManagement'),
		(?, ?, 'encodeAssets')
	`, userID, companyID, userID, companyID, userID, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to create user roles: " + err.Error(),
		})
		return
	}

	// Create default asset categories
	defaultCategories := []struct {
		name  string
		color string
	}{
		{"Computer", "#007bff"},
		{"Printer", "#28a745"},
		{"Network", "#ffc107"},
		{"Furniture", "#dc3545"},
		{"Vehicle", "#6f42c1"},
		{"Equipment", "#fd7e14"},
	}

	for _, cat := range defaultCategories {
		_, err = tx.Exec(`
			INSERT INTO asset_categories (company_id, name, description, color)
			VALUES (?, ?, ?, ?)
		`, companyID, cat.name, cat.name+" equipment", cat.color)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   "Failed to create default categories: " + err.Error(),
			})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Company registered successfully",
		Data: map[string]interface{}{
			"company_id":   companyID,
			"user_id":      userID,
			"company_code": req.CompanyCode,
		},
	})
}

// authMiddleware validates JWT token and adds user info to context
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Error:   "Authorization header required",
			})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			secretKey = "your-secret-key"
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Error:   "Invalid or expired token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Error:   "Invalid token claims",
			})
			c.Abort()
			return
		}

		// Check if user still exists and is active
		var user User
		err = db.QueryRow(`
			SELECT id, company_id, username, email, first_name, last_name, role, is_active
			FROM users 
			WHERE id = ? AND company_id = ? AND is_active = true
		`, claims.UserID, claims.CompanyID).Scan(
			&user.ID, &user.CompanyID, &user.Username, &user.Email,
			&user.FirstName, &user.LastName, &user.Role, &user.IsActive,
		)

		if err != nil {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Error:   "User not found or inactive",
			})
			c.Abort()
			return
		}

		// Add user info to context
		c.Set("user", user)
		c.Set("company_id", claims.CompanyID)
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// checkTrialStatus checks if company trial is still active
func checkTrialStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		companyID := c.GetInt("company_id")

		var trialEndsAt *time.Time
		err := db.QueryRow("SELECT trial_ends_at FROM companies WHERE id = ?", companyID).Scan(&trialEndsAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   "Failed to check trial status",
			})
			c.Abort()
			return
		}

		if trialEndsAt != nil && time.Now().After(*trialEndsAt) {
			c.JSON(http.StatusForbidden, APIResponse{
				Success: false,
				Error:   "Trial period has expired. Please contact support to upgrade.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getCurrentUser returns the current authenticated user
func getCurrentUser(c *gin.Context) *User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	if u, ok := user.(User); ok {
		return &u
	}
	return nil
}

// getCurrentCompanyID returns the current company ID
func getCurrentCompanyID(c *gin.Context) int {
	companyID, exists := c.Get("company_id")
	if !exists {
		return 0
	}
	if id, ok := companyID.(int); ok {
		return id
	}
	return 0
} 