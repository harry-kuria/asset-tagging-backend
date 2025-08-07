package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	trialPeriodDays = 30
)

// Claims represents JWT claims
type Claims struct {
	UserID            int    `json:"userId"`
	Username          string `json:"username"`
	IsLicenseActive   bool   `json:"isLicenseActive"`
	TrialStartDate    string `json:"trialStartDate"`
	jwt.RegisteredClaims
}

// generateToken creates a JWT token for the user
func generateToken(user User) (string, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "your-secret-key" // Default fallback
	}

	claims := Claims{
		UserID:          user.ID,
		Username:        user.Username,
		IsLicenseActive: user.IsLicenseActive,
		TrialStartDate:  user.TrialStartDate.Format("2006-01-02"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPassword compares a password with its hash
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateRandomPassword generates a random password
func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	
	for i := range password {
		randomByte := make([]byte, 1)
		_, err := rand.Read(randomByte)
		if err != nil {
			return "", err
		}
		password[i] = charset[randomByte[0]%byte(len(charset))]
	}
	
	return string(password), nil
}

// loginHandler handles user login
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request data",
		})
		return
	}

	log.Printf("Received login request for username: %s", req.Username)

	// Query the database for the user
	var user User
	err := db.QueryRow("SELECT id, username, password, trialStartDate, isLicenseActive, createdAt, updatedAt FROM users WHERE username = ?", req.Username).
		Scan(&user.ID, &user.Username, &user.Password, &user.TrialStartDate, &user.IsLicenseActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Error:   "Invalid username or password",
			})
			return
		}
		log.Printf("Error executing MySQL query: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	// Check password
	if !checkPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Error:   "Invalid username or password",
		})
		return
	}

	// Generate JWT token
	token, err := generateToken(user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	// Get user roles
	var userRole UserRole
	err = db.QueryRow("SELECT id, user_id, userManagement, assetManagement, encodeAssets, addMultipleAssets, viewReports, printReports FROM user_roles WHERE user_id = ?", user.ID).
		Scan(&userRole.ID, &userRole.UserID, &userRole.UserManagement, &userRole.AssetManagement, &userRole.EncodeAssets, &userRole.AddMultipleAssets, &userRole.ViewReports, &userRole.PrintReports)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error fetching user roles: %v", err)
	}

	response := gin.H{
		"success": true,
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":              user.ID,
			"username":        user.Username,
			"trialStartDate":  user.TrialStartDate,
			"isLicenseActive": user.IsLicenseActive,
			"roles":           userRole,
		},
	}

	c.JSON(http.StatusOK, response)
}

// createAccountHandler handles account creation
func createAccountHandler(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request data",
		})
		return
	}

	// Check if username already exists
	var existingUser int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Error:   "Username already exists",
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

	// Insert new user
	result, err := db.Exec("INSERT INTO users (username, password, trialStartDate, isLicenseActive, createdAt, updatedAt) VALUES (?, ?, ?, ?, ?, ?)",
		req.Username, hashedPassword, time.Now(), false, time.Now(), time.Now())
	if err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	userID, _ := result.LastInsertId()

	// Insert default user roles
	_, err = db.Exec("INSERT INTO user_roles (user_id, userManagement, assetManagement, encodeAssets, addMultipleAssets, viewReports, printReports) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, true, true, true, true, true, true)
	if err != nil {
		log.Printf("Error creating user roles: %v", err)
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Account created successfully",
		Data: gin.H{
			"userId": userID,
		},
	})
}

// checkTrialStatus middleware checks if the user's trial is still valid
func checkTrialStatus(c *gin.Context) {
	// Get user from JWT token
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		c.Abort()
		return
	}

	userData := user.(User)

	// If the user has an active license, skip trial check
	if userData.IsLicenseActive {
		c.Next()
		return
	}

	// Calculate the trial end date and check if it's expired
	trialEndDate := userData.TrialStartDate.AddDate(0, 0, trialPeriodDays)
	currentDate := time.Now()

	if currentDate.After(trialEndDate) {
		c.JSON(http.StatusForbidden, APIResponse{
			Success: false,
			Error:   "Trial period has expired. Please purchase a license.",
		})
		c.Abort()
		return
	}

	c.Next()
}

// authMiddleware validates JWT token and sets user in context
func authMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Error:   "Authorization header required",
		})
		c.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Error:   "Bearer token required",
		})
		c.Abort()
		return
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
			Error:   "Invalid token",
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

	// Get user from database
	var user User
	err = db.QueryRow("SELECT id, username, password, trialStartDate, isLicenseActive, createdAt, updatedAt FROM users WHERE id = ?", claims.UserID).
		Scan(&user.ID, &user.Username, &user.Password, &user.TrialStartDate, &user.IsLicenseActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Error:   "User not found",
		})
		c.Abort()
		return
	}

	c.Set("user", user)
	c.Next()
}

// numberToWords converts a number to words
func numberToWords(num int) string {
	if num == 0 {
		return "zero"
	}

	units := []string{"", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}
	teens := []string{"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}
	tens := []string{"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}

	if num < 10 {
		return units[num]
	} else if num < 20 {
		return teens[num-10]
	} else if num < 100 {
		if num%10 == 0 {
			return tens[num/10]
		}
		return tens[num/10] + " " + units[num%10]
	} else if num < 1000 {
		if num%100 == 0 {
			return units[num/100] + " hundred"
		}
		return units[num/100] + " hundred and " + numberToWords(num%100)
	} else if num < 1000000 {
		if num%1000 == 0 {
			return numberToWords(num/1000) + " thousand"
		}
		return numberToWords(num/1000) + " thousand " + numberToWords(num%1000)
	}

	return fmt.Sprintf("%d", num)
}

// formatCurrency formats a float as currency
func formatCurrency(amount float64) string {
	return fmt.Sprintf("₱%.2f", amount)
} 