package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// getUsersHandler returns all users for the current company
func getUsersHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)

	rows, err := db.Query(`
		SELECT id, company_id, username, email, first_name, last_name, role, is_active, last_login, created_at, updated_at
		FROM users 
		WHERE company_id = ? AND is_active = true
		ORDER BY username
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to fetch users: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID, &user.CompanyID, &user.Username, &user.Email,
			&user.FirstName, &user.LastName, &user.Role, &user.IsActive,
			&user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   "Failed to scan user: " + err.Error(),
			})
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    users,
	})
}

// getUserHandler returns a specific user by ID
func getUserHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	var user User
	err = db.QueryRow(`
		SELECT id, company_id, username, email, first_name, last_name, role, is_active, last_login, created_at, updated_at
		FROM users 
		WHERE id = ? AND company_id = ? AND is_active = true
	`, userID, companyID).Scan(
		&user.ID, &user.CompanyID, &user.Username, &user.Email,
		&user.FirstName, &user.LastName, &user.Role, &user.IsActive,
		&user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Error:   "User not found",
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
		Data:    user,
	})
}

// addUserHandler adds a new user to the current company
func addUserHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)

	var req AddUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Check if username already exists in this company
	var existingID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ? AND company_id = ?", req.Username, companyID).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Error:   "Username already exists",
		})
		return
	}

	// Check if email already exists in this company
	err = db.QueryRow("SELECT id FROM users WHERE email = ? AND company_id = ?", req.Email, companyID).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Error:   "Email already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to hash password",
		})
		return
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = "user"
	}

	// Insert the user
	result, err := db.Exec(`
		INSERT INTO users (company_id, username, email, password_hash, first_name, last_name, role)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, companyID, req.Username, req.Email, string(hashedPassword), req.FirstName, req.LastName, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to create user: " + err.Error(),
		})
		return
	}

	userID, _ := result.LastInsertId()

	// Create default user roles based on the role
	var roles []string
	switch req.Role {
	case "admin":
		roles = []string{"userManagement", "assetManagement", "encodeAssets"}
	case "manager":
		roles = []string{"assetManagement", "encodeAssets"}
	default:
		roles = []string{"encodeAssets"}
	}

	// Insert user roles
	for _, role := range roles {
		_, err = db.Exec("INSERT INTO user_roles (user_id, company_id, role) VALUES (?, ?, ?)", userID, companyID, role)
		if err != nil {
			log.Printf("Error adding user role %s: %v", role, err)
		}
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "User added successfully",
		Data: map[string]interface{}{
			"user_id": userID,
		},
	})
}

// updateUserHandler updates an existing user
func updateUserHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Check if user exists and belongs to this company
	var existingID int
	err = db.QueryRow("SELECT id FROM users WHERE id = ? AND company_id = ?", userID, companyID).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Error:   "User not found",
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
	query := "UPDATE users SET "
	var params []interface{}
	var updates []string

	if req.Username != "" {
		updates = append(updates, "username = ?")
		params = append(params, req.Username)
	}
	if req.Email != "" {
		updates = append(updates, "email = ?")
		params = append(params, req.Email)
	}
	if req.FirstName != "" {
		updates = append(updates, "first_name = ?")
		params = append(params, req.FirstName)
	}
	if req.LastName != "" {
		updates = append(updates, "last_name = ?")
		params = append(params, req.LastName)
	}
	if req.Role != "" {
		updates = append(updates, "role = ?")
		params = append(params, req.Role)
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
	params = append(params, userID, companyID)

	_, err = db.Exec(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to update user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "User updated successfully",
	})
}

// deleteUserHandler deletes a user (soft delete by setting is_active = false)
func deleteUserHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	// Check if user exists and belongs to this company
	var existingID int
	err = db.QueryRow("SELECT id FROM users WHERE id = ? AND company_id = ?", userID, companyID).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Error:   "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database error: " + err.Error(),
		})
		return
	}

	// Soft delete by setting is_active = false
	_, err = db.Exec("UPDATE users SET is_active = false, updated_at = NOW() WHERE id = ? AND company_id = ?", userID, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to delete user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "User deleted successfully",
	})
} 