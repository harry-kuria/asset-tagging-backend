package main

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID             int       `json:"id" db:"id"`
	Username       string    `json:"username" db:"username"`
	Password       string    `json:"password" db:"password"`
	TrialStartDate time.Time `json:"trialStartDate" db:"trialStartDate"`
	IsLicenseActive bool     `json:"isLicenseActive" db:"isLicenseActive"`
	CreatedAt      time.Time `json:"createdAt" db:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updatedAt"`
}

// UserRole represents user permissions
type UserRole struct {
	ID                int `json:"id" db:"id"`
	UserID            int `json:"userId" db:"user_id"`
	UserManagement    bool `json:"userManagement" db:"userManagement"`
	AssetManagement   bool `json:"assetManagement" db:"assetManagement"`
	EncodeAssets      bool `json:"encodeAssets" db:"encodeAssets"`
	AddMultipleAssets bool `json:"addMultipleAssets" db:"addMultipleAssets"`
	ViewReports       bool `json:"viewReports" db:"viewReports"`
	PrintReports      bool `json:"printReports" db:"printReports"`
}

// Asset represents an asset in the system
type Asset struct {
	ID              int       `json:"id" db:"id"`
	AssetName       string    `json:"assetName" db:"assetName"`
	AssetType       string    `json:"assetType" db:"assetType"`
	InstitutionName string    `json:"institutionName" db:"institutionName"`
	Department      string    `json:"department" db:"department"`
	FunctionalArea  string    `json:"functionalArea" db:"functionalArea"`
	Manufacturer    string    `json:"manufacturer" db:"manufacturer"`
	ModelNumber     string    `json:"modelNumber" db:"modelNumber"`
	SerialNumber    string    `json:"serialNumber" db:"serialNumber"`
	Location        string    `json:"location" db:"location"`
	Status          string    `json:"status" db:"status"`
	PurchaseDate    time.Time `json:"purchaseDate" db:"purchaseDate"`
	PurchasePrice   float64   `json:"purchasePrice" db:"purchasePrice"`
	Logo            string    `json:"logo" db:"logo"`
	CreatedAt       time.Time `json:"createdAt" db:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt" db:"updatedAt"`
}

// AssetCategory represents asset categories
type AssetCategory struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

// LoginRequest represents login request data
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateAccountRequest represents account creation request
type CreateAccountRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AddUserRequest represents user creation request
type AddUserRequest struct {
	Username string   `json:"username" binding:"required"`
	Password string   `json:"password" binding:"required"`
	Roles    []string `json:"roles" binding:"required"`
}

// AssetRequest represents asset creation/update request
type AssetRequest struct {
	AssetName       string  `json:"assetName" binding:"required"`
	AssetType       string  `json:"assetType" binding:"required"`
	InstitutionName string  `json:"institutionName"`
	Department      string  `json:"department"`
	FunctionalArea  string  `json:"functionalArea"`
	Manufacturer    string  `json:"manufacturer"`
	ModelNumber     string  `json:"modelNumber"`
	SerialNumber    string  `json:"serialNumber"`
	Location        string  `json:"location"`
	Status          string  `json:"status"`
	PurchaseDate    string  `json:"purchaseDate"`
	PurchasePrice   float64 `json:"purchasePrice"`
}

// MultipleAssetRequest represents multiple asset creation request
type MultipleAssetRequest struct {
	Assets []AssetRequest `json:"assets" binding:"required"`
}

// BarcodeRequest represents barcode generation request
type BarcodeRequest struct {
	AssetIDs []int `json:"assetIds" binding:"required"`
}

// InstitutionBarcodeRequest represents institution-based barcode generation
type InstitutionBarcodeRequest struct {
	Institution string `json:"institution" binding:"required"`
}

// InstitutionDepartmentBarcodeRequest represents institution and department barcode generation
type InstitutionDepartmentBarcodeRequest struct {
	Institution string `json:"institution" binding:"required"`
	Department  string `json:"department" binding:"required"`
}

// ReportRequest represents report generation request
type ReportRequest struct {
	AssetType       string   `json:"assetType"`
	Location        string   `json:"location"`
	Status          string   `json:"status"`
	StartDate       string   `json:"startDate"`
	EndDate         string   `json:"endDate"`
	Manufacturer    []string `json:"manufacturer"`
	ModelNumber     string   `json:"modelNumber"`
	InstitutionName string   `json:"institutionName"`
	Department      string   `json:"department"`
	FunctionalArea  string   `json:"functionalArea"`
}

// InvoiceRequest represents invoice generation request
type InvoiceRequest struct {
	CustomerName    string  `json:"customerName" binding:"required"`
	CustomerAddress string  `json:"customerAddress" binding:"required"`
	Items           []InvoiceItem `json:"items" binding:"required"`
}

// InvoiceItem represents an item in an invoice
type InvoiceItem struct {
	Description string  `json:"description" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required"`
	UnitPrice   float64 `json:"unitPrice" binding:"required"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SearchRequest represents asset search request
type SearchRequest struct {
	Query string `json:"query" binding:"required"`
}

// AssetDetailsRequest represents asset details request
type AssetDetailsRequest struct {
	AssetID int `json:"assetId" binding:"required"`
} 