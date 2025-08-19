package main

import (
	"time"
)

// Company represents a company/organization
type Company struct {
	ID               int       `json:"id" db:"id"`
	CompanyName      string    `json:"company_name" db:"company_name"`
	CompanyCode      string    `json:"company_code" db:"company_code"`
	Email            string    `json:"email" db:"email"`
	Phone            *string   `json:"phone" db:"phone"`
	Address          *string   `json:"address" db:"address"`
	Industry         *string   `json:"industry" db:"industry"`
	SubscriptionPlan string    `json:"subscription_plan" db:"subscription_plan"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	TrialEndsAt      *time.Time `json:"trial_ends_at" db:"trial_ends_at"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// User represents a user with company association
type User struct {
	ID           int       `json:"id" db:"id"`
	CompanyID    int       `json:"company_id" db:"company_id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FirstName    *string   `json:"first_name" db:"first_name"`
	LastName     *string   `json:"last_name" db:"last_name"`
	Role         string    `json:"role" db:"role"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	LastLogin    *time.Time `json:"last_login" db:"last_login"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole represents user roles with company association
type UserRole struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	CompanyID int       `json:"company_id" db:"company_id"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AssetCategory represents asset categories with company association
type AssetCategory struct {
	ID          int       `json:"id" db:"id"`
	CompanyID   int       `json:"company_id" db:"company_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	Color       *string   `json:"color" db:"color"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Asset represents an asset with company association
type Asset struct {
	ID               int       `json:"id" db:"id"`
	CompanyID        int       `json:"company_id" db:"company_id"`
	AssetName        string    `json:"asset_name" db:"asset_name"`
	AssetType        *string   `json:"asset_type" db:"asset_type"`
	CategoryID       *int      `json:"category_id" db:"category_id"`
	InstitutionName  *string   `json:"institution_name" db:"institution_name"`
	Department       *string   `json:"department" db:"department"`
	FunctionalArea   *string   `json:"functional_area" db:"functional_area"`
	Manufacturer     *string   `json:"manufacturer" db:"manufacturer"`
	ModelNumber      *string   `json:"model_number" db:"model_number"`
	SerialNumber     *string   `json:"serial_number" db:"serial_number"`
	Location         *string   `json:"location" db:"location"`
	Status           string    `json:"status" db:"status"`
	PurchaseDate     *time.Time `json:"purchase_date" db:"purchase_date"`
	PurchasePrice    *float64  `json:"purchase_price" db:"purchase_price"`
	AssignedTo       *int      `json:"assigned_to" db:"assigned_to"`
	Notes            *string   `json:"notes" db:"notes"`
	Barcode          *string   `json:"barcode" db:"barcode"`
	QRCode           *string   `json:"qr_code" db:"qr_code"`
	Logo             *string   `json:"logo" db:"logo"`
	CreatedBy        int       `json:"created_by" db:"created_by"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// AssetMaintenance represents asset maintenance history
type AssetMaintenance struct {
	ID                   int       `json:"id" db:"id"`
	CompanyID            int       `json:"company_id" db:"company_id"`
	AssetID              int       `json:"asset_id" db:"asset_id"`
	MaintenanceType      string    `json:"maintenance_type" db:"maintenance_type"`
	Description          string    `json:"description" db:"description"`
	Cost                 *float64  `json:"cost" db:"cost"`
	PerformedBy          *string   `json:"performed_by" db:"performed_by"`
	PerformedAt          time.Time `json:"performed_at" db:"performed_at"`
	NextMaintenanceDate  *time.Time `json:"next_maintenance_date" db:"next_maintenance_date"`
	CreatedBy            int       `json:"created_by" db:"created_by"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

// AssetAssignment represents asset assignment history
type AssetAssignment struct {
	ID          int       `json:"id" db:"id"`
	CompanyID   int       `json:"company_id" db:"company_id"`
	AssetID     int       `json:"asset_id" db:"asset_id"`
	AssignedTo  int       `json:"assigned_to" db:"assigned_to"`
	AssignedBy  int       `json:"assigned_by" db:"assigned_by"`
	AssignedAt  time.Time `json:"assigned_at" db:"assigned_at"`
	ReturnedAt  *time.Time `json:"returned_at" db:"returned_at"`
	Notes       *string   `json:"notes" db:"notes"`
}

// CompanySetting represents company settings
type CompanySetting struct {
	ID         int       `json:"id" db:"id"`
	CompanyID  int       `json:"company_id" db:"company_id"`
	SettingKey string    `json:"setting_key" db:"setting_key"`
	SettingValue *string `json:"setting_value" db:"setting_value"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Request/Response structures

// LoginRequest represents login request
type LoginRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CompanyCode string `json:"company_code"` // optional; if empty, login resolves by username only
}

// RegisterCompanyRequest represents company registration request
type RegisterCompanyRequest struct {
	CompanyName string              `json:"company_name" binding:"required"`
	CompanyCode string              `json:"company_code"` // optional; server will auto-generate if empty
	Email       string              `json:"email" binding:"required,email"`
	Phone       string              `json:"phone"`
	Address     string              `json:"address"`
	Industry    string              `json:"industry"`
	AdminUser   RegisterUserRequest `json:"admin_user" binding:"required"`
}

// RegisterUserRequest represents user registration request
type RegisterUserRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// AddUserRequest represents adding a user to a company
type AddUserRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	IsActive  *bool  `json:"is_active"`
}

// AddAssetRequest represents adding an asset
type AddAssetRequest struct {
	AssetName       string  `json:"asset_name" binding:"required"`
	AssetType       string  `json:"asset_type"`
	CategoryID      *int    `json:"category_id"`
	InstitutionName string  `json:"institution_name"`
	Department      string  `json:"department"`
	FunctionalArea  string  `json:"functional_area"`
	Manufacturer    string  `json:"manufacturer"`
	ModelNumber     string  `json:"model_number"`
	SerialNumber    string  `json:"serial_number"`
	Location        string  `json:"location"`
	Status          string  `json:"status"`
	PurchaseDate    string  `json:"purchase_date"`
	PurchasePrice   *float64 `json:"purchase_price"`
	AssignedTo      *int    `json:"assigned_to"`
	Notes           string  `json:"notes"`
}

// UpdateAssetRequest represents updating an asset
type UpdateAssetRequest struct {
	AssetName       string  `json:"asset_name"`
	AssetType       string  `json:"asset_type"`
	CategoryID      *int    `json:"category_id"`
	InstitutionName string  `json:"institution_name"`
	Department      string  `json:"department"`
	FunctionalArea  string  `json:"functional_area"`
	Manufacturer    string  `json:"manufacturer"`
	ModelNumber     string  `json:"model_number"`
	SerialNumber    string  `json:"serial_number"`
	Location        string  `json:"location"`
	Status          string  `json:"status"`
	PurchaseDate    string  `json:"purchase_date"`
	PurchasePrice   *float64 `json:"purchase_price"`
	AssignedTo      *int    `json:"assigned_to"`
	Notes           string  `json:"notes"`
}

// SearchAssetsRequest represents asset search request
type SearchAssetsRequest struct {
	Query           string `json:"query"`
	AssetType       string `json:"asset_type"`
	InstitutionName string `json:"institution_name"`
	Department      string `json:"department"`
	Status          string `json:"status"`
	CategoryID      *int   `json:"category_id"`
}

// AssetRequest represents asset creation/update request (legacy compatibility)
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

// AddCategoryRequest represents adding an asset category
type AddCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// UpdateCategoryRequest represents updating an asset category
type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	IsActive    *bool  `json:"is_active"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token     string  `json:"token"`
	User      User    `json:"user"`
	Company   Company `json:"company"`
	ExpiresAt int64   `json:"expires_at"`
	Roles     []string `json:"roles"`
}

// APIResponse represents generic API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse represents paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalAssets     int     `json:"total_assets"`
	ActiveAssets    int     `json:"active_assets"`
	TotalUsers      int     `json:"total_users"`
	TotalValue      float64 `json:"total_value"`
	TotalBarcodes   int     `json:"total_barcodes"`
	ScannedBarcodes int     `json:"scanned_barcodes"`
	AssetsByStatus  map[string]int `json:"assets_by_status"`
	AssetsByType    map[string]int `json:"assets_by_type"`
	RecentAssets    []Asset `json:"recent_assets"`
	RecentMaintenance []AssetMaintenance `json:"recent_maintenance"`
}