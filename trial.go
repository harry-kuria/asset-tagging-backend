package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TrialStatus represents the trial status information
type TrialStatus struct {
	IsActive        bool      `json:"is_active"`
	DaysRemaining   int       `json:"days_remaining"`
	TrialEndsAt     time.Time `json:"trial_ends_at"`
	IsExpired       bool      `json:"is_expired"`
	SubscriptionPlan string   `json:"subscription_plan"`
	RequiresPayment bool      `json:"requires_payment"`
}

// PaymentPlan represents available subscription plans
type PaymentPlan struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	BillingCycle string `json:"billing_cycle"`
	Features    []string `json:"features"`
}

// getTrialStatusHandler returns current trial status for the company
func getTrialStatusHandler(c *gin.Context) {
	companyID := getCurrentCompanyID(c)

	var trialEndsAt *time.Time
	var subscriptionPlan string
	var isActive bool

	err := db.QueryRow(`
		SELECT trial_ends_at, subscription_plan, is_active 
		FROM companies WHERE id = ?
	`, companyID).Scan(&trialEndsAt, &subscriptionPlan, &isActive)

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get trial status: " + err.Error(),
		})
		return
	}

	now := time.Now()
	var trialStatus TrialStatus

	if trialEndsAt == nil {
		// No trial period set
		trialStatus = TrialStatus{
			IsActive:        false,
			DaysRemaining:   0,
			TrialEndsAt:     now,
			IsExpired:       true,
			SubscriptionPlan: subscriptionPlan,
			RequiresPayment: true,
		}
	} else {
		daysRemaining := int(trialEndsAt.Sub(now).Hours() / 24)
		isExpired := now.After(*trialEndsAt)

		trialStatus = TrialStatus{
			IsActive:        isActive && !isExpired,
			DaysRemaining:   daysRemaining,
			TrialEndsAt:     *trialEndsAt,
			IsExpired:       isExpired,
			SubscriptionPlan: subscriptionPlan,
			RequiresPayment: isExpired || subscriptionPlan == "trial",
		}
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    trialStatus,
	})
}

// getPaymentPlansHandler returns available subscription plans
func getPaymentPlansHandler(c *gin.Context) {
	plans := []PaymentPlan{
		{
			ID:          "basic",
			Name:        "Basic Plan",
			Price:       29.99,
			Currency:    "USD",
			BillingCycle: "monthly",
			Features: []string{
				"Up to 100 assets",
				"Basic reporting",
				"Email support",
				"Barcode generation",
			},
		},
		{
			ID:          "professional",
			Name:        "Professional Plan",
			Price:       79.99,
			Currency:    "USD",
			BillingCycle: "monthly",
			Features: []string{
				"Up to 1000 assets",
				"Advanced reporting",
				"Priority support",
				"Custom branding",
				"API access",
				"Bulk operations",
			},
		},
		{
			ID:          "enterprise",
			Name:        "Enterprise Plan",
			Price:       199.99,
			Currency:    "USD",
			BillingCycle: "monthly",
			Features: []string{
				"Unlimited assets",
				"Custom integrations",
				"Dedicated support",
				"Advanced analytics",
				"Multi-location support",
				"Custom workflows",
			},
		},
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    plans,
	})
}

// initiatePaymentHandler starts the payment process
func initiatePaymentHandler(c *gin.Context) {
	var req struct {
		PlanID string `json:"plan_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	companyID := getCurrentCompanyID(c)

	// Validate plan ID
	validPlans := map[string]bool{"basic": true, "professional": true, "enterprise": true}
	if !validPlans[req.PlanID] {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid plan ID",
		})
		return
	}

	// Get company information
	var companyName, companyCode string
	err := db.QueryRow("SELECT company_name, company_code FROM companies WHERE id = ?", companyID).Scan(&companyName, &companyCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get company information",
		})
		return
	}

	// Create payment session (in a real implementation, this would integrate with Stripe/PayPal)
	paymentSession := map[string]interface{}{
		"session_id":    generatePaymentSessionID(),
		"company_id":    companyID,
		"company_name":  companyName,
		"company_code":  companyCode,
		"plan_id":       req.PlanID,
		"amount":        getPlanPrice(req.PlanID),
		"currency":      "USD",
		"expires_at":    time.Now().Add(30 * time.Minute),
		"payment_url":   "https://payment.moowigroup.com/checkout", // Replace with actual payment URL
		"webhook_url":   "https://graf.moowigroup.com/api/payment/webhook",
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Payment session created successfully",
		Data:    paymentSession,
	})
}

// paymentWebhookHandler handles payment completion webhooks
func paymentWebhookHandler(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Status    string `json:"status" binding:"required"`
		PlanID    string `json:"plan_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid webhook data",
		})
		return
	}

	// In a real implementation, verify the webhook signature
	// For now, we'll trust the webhook data

	if req.Status == "completed" {
		// Update company subscription
		_, err := db.Exec(`
			UPDATE companies 
			SET subscription_plan = ?, 
				is_active = true, 
				trial_ends_at = NULL,
				updated_at = NOW()
			WHERE id = (SELECT company_id FROM payment_sessions WHERE session_id = ?)
		`, req.PlanID, req.SessionID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   "Failed to update subscription",
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Subscription updated successfully",
		})
	} else {
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Payment not completed",
		})
	}
}

// checkTrialStatusMiddleware enhanced trial status checker
func checkTrialStatusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		companyID := c.GetInt("company_id")

		var trialEndsAt *time.Time
		var subscriptionPlan string
		var isActive bool

		err := db.QueryRow(`
			SELECT trial_ends_at, subscription_plan, is_active 
			FROM companies WHERE id = ?
		`, companyID).Scan(&trialEndsAt, &subscriptionPlan, &isActive)

		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   "Failed to check trial status",
			})
			c.Abort()
			return
		}

		// Check if trial has expired
		if trialEndsAt != nil && time.Now().After(*trialEndsAt) {
			c.JSON(http.StatusForbidden, APIResponse{
				Success: false,
				Error:   "Trial period has expired. Please upgrade your subscription to continue.",
				Data: map[string]interface{}{
					"trial_expired": true,
					"requires_payment": true,
					"subscription_plan": subscriptionPlan,
				},
			})
			c.Abort()
			return
		}

		// Check if company is active
		if !isActive {
			c.JSON(http.StatusForbidden, APIResponse{
				Success: false,
				Error:   "Account is suspended. Please contact support.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper functions
func generatePaymentSessionID() string {
	return "sess_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func getPlanPrice(planID string) float64 {
	prices := map[string]float64{
		"basic":        29.99,
		"professional": 79.99,
		"enterprise":   199.99,
	}
	return prices[planID]
} 