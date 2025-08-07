package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// getUsersHandler returns all users
func getUsersHandler(c *gin.Context) {
	rows, err := db.Query("SELECT id, username, trialStartDate, isLicenseActive, createdAt, updatedAt FROM users")
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}
	defer rows.Close()

	var users []gin.H
	for rows.Next() {
		var user struct {
			ID             int    `json:"id"`
			Username       string `json:"username"`
			TrialStartDate string `json:"trialStartDate"`
			IsLicenseActive bool  `json:"isLicenseActive"`
			CreatedAt      string `json:"createdAt"`
			UpdatedAt      string `json:"updatedAt"`
		}

		err := rows.Scan(&user.ID, &user.Username, &user.TrialStartDate, &user.IsLicenseActive, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}

		users = append(users, gin.H{
			"id":              user.ID,
			"username":        user.Username,
			"trialStartDate":  user.TrialStartDate,
			"isLicenseActive": user.IsLicenseActive,
			"createdAt":       user.CreatedAt,
			"updatedAt":       user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, users)
}

// getUserHandler returns a specific user by ID
func getUserHandler(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	var user struct {
		ID             int    `json:"id"`
		Username       string `json:"username"`
		TrialStartDate string `json:"trialStartDate"`
		IsLicenseActive bool  `json:"isLicenseActive"`
		CreatedAt      string `json:"createdAt"`
		UpdatedAt      string `json:"updatedAt"`
	}

	err = db.QueryRow("SELECT id, username, trialStartDate, isLicenseActive, createdAt, updatedAt FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Username, &user.TrialStartDate, &user.IsLicenseActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Error:   "User not found",
			})
			return
		}
		log.Printf("Error fetching user: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// addUserHandler creates a new user
func addUserHandler(c *gin.Context) {
	var req AddUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	// Insert the new user
	result, err := db.Exec("INSERT INTO users (username, password, trialStartDate, isLicenseActive, createdAt, updatedAt) VALUES (?, ?, ?, ?, ?, ?)",
		req.Username, hashedPassword, time.Now(), false, time.Now(), time.Now())
	if err != nil {
		log.Printf("Error adding user: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	userID, _ := result.LastInsertId()

	// Insert the roles for the user
	_, err = db.Exec("INSERT INTO user_roles (user_id, userManagement, assetManagement, encodeAssets, addMultipleAssets, viewReports, printReports) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID,
		contains(req.Roles, "userManagement"),
		contains(req.Roles, "assetManagement"),
		contains(req.Roles, "encodeAssets"),
		contains(req.Roles, "addMultipleAssets"),
		contains(req.Roles, "viewReports"),
		contains(req.Roles, "printReports"))
	if err != nil {
		log.Printf("Error adding user roles: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "User added successfully",
		Data: gin.H{
			"userId": userID,
		},
	})
}

// updateUserHandler updates an existing user
func updateUserHandler(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	var req AddUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid input data",
		})
		return
	}

	// Hash password if provided
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	// Update the user
	_, err = db.Exec("UPDATE users SET username = ?, password = ?, updatedAt = ? WHERE id = ?",
		req.Username, hashedPassword, time.Now(), userID)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	// Update user roles
	_, err = db.Exec("UPDATE user_roles SET userManagement = ?, assetManagement = ?, encodeAssets = ?, addMultipleAssets = ?, viewReports = ?, printReports = ? WHERE user_id = ?",
		contains(req.Roles, "userManagement"),
		contains(req.Roles, "assetManagement"),
		contains(req.Roles, "encodeAssets"),
		contains(req.Roles, "addMultipleAssets"),
		contains(req.Roles, "viewReports"),
		contains(req.Roles, "printReports"),
		userID)
	if err != nil {
		log.Printf("Error updating user roles: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "User updated successfully",
	})
}

// deleteUserHandler deletes a user
func deleteUserHandler(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	// Delete user roles first
	_, err = db.Exec("DELETE FROM user_roles WHERE user_id = ?", userID)
	if err != nil {
		log.Printf("Error deleting user roles: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	// Delete the user
	_, err = db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
} 