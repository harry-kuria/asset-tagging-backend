-- Asset Tagging Database Schema with Company Multi-tenancy
-- Each company has isolated data

CREATE DATABASE IF NOT EXISTS asset_management;
USE asset_management;

-- Companies table (top-level organization)
CREATE TABLE IF NOT EXISTS companies (
    id INT AUTO_INCREMENT PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    company_code VARCHAR(50) UNIQUE NOT NULL, -- Unique identifier for company
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50),
    address TEXT,
    industry VARCHAR(100),
    subscription_plan VARCHAR(50) DEFAULT 'basic',
    is_active BOOLEAN DEFAULT TRUE,
    trial_ends_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Users table with company association
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    company_id INT NOT NULL,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role ENUM('admin', 'manager', 'user') DEFAULT 'user',
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    UNIQUE KEY unique_username_per_company (company_id, username),
    UNIQUE KEY unique_email_per_company (company_id, email)
);

-- User roles with company association
CREATE TABLE IF NOT EXISTS user_roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    company_id INT NOT NULL,
    role VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
);

-- Asset categories with company association
CREATE TABLE IF NOT EXISTS asset_categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    company_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#007bff', -- Hex color for UI
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    UNIQUE KEY unique_category_per_company (company_id, name)
);
someone
-- Assets table with company association
CREATE TABLE IF NOT EXISTS assets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    company_id INT NOT NULL,
    asset_name VARCHAR(255) NOT NULL,
    asset_type VARCHAR(100),
    category_id INT,
    institution_name VARCHAR(255),
    department VARCHAR(255),
    functional_area VARCHAR(255),
    manufacturer VARCHAR(255),
    model_number VARCHAR(255),
    serial_number VARCHAR(255),
    location VARCHAR(255),
    status ENUM('Active', 'Inactive', 'Maintenance', 'Retired') DEFAULT 'Active',
    purchase_date DATE,
    purchase_price DECIMAL(10,2),
    assigned_to INT NULL,
    notes TEXT,
    barcode VARCHAR(255) UNIQUE,
    qr_code VARCHAR(255) UNIQUE,
    created_by INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES asset_categories(id) ON DELETE SET NULL,
    FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Asset maintenance history
CREATE TABLE IF NOT EXISTS asset_maintenance (
    id INT AUTO_INCREMENT PRIMARY KEY,
    company_id INT NOT NULL,
    asset_id INT NOT NULL,
    maintenance_type ENUM('Preventive', 'Corrective', 'Emergency') NOT NULL,
    description TEXT NOT NULL,
    cost DECIMAL(10,2),
    performed_by VARCHAR(255),
    performed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    next_maintenance_date DATE,
    created_by INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Asset assignments history
CREATE TABLE IF NOT EXISTS asset_assignments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    company_id INT NOT NULL,
    asset_id INT NOT NULL,
    assigned_to INT NOT NULL,
    assigned_by INT NOT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    returned_at TIMESTAMP NULL,
    notes TEXT,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Company settings
CREATE TABLE IF NOT EXISTS company_settings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    company_id INT NOT NULL,
    setting_key VARCHAR(100) NOT NULL,
    setting_value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    UNIQUE KEY unique_setting_per_company (company_id, setting_key)
);

-- Insert default company (for existing data migration)
INSERT IGNORE INTO companies (id, company_name, company_code, email, industry) VALUES 
(1, 'Default Company', 'DEFAULT', 'admin@default.com', 'Technology');

-- Insert default admin user
INSERT IGNORE INTO users (id, company_id, username, email, password_hash, first_name, last_name, role) VALUES 
(1, 1, 'admin', 'admin@default.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Admin', 'User', 'admin');

-- Insert default user roles
INSERT IGNORE INTO user_roles (user_id, company_id, role) VALUES 
(1, 1, 'userManagement'), 
(1, 1, 'assetManagement'), 
(1, 1, 'encodeAssets');

-- Insert default asset categories
INSERT IGNORE INTO asset_categories (company_id, name, description, color) VALUES 
(1, 'Computer', 'Desktop and laptop computers', '#007bff'),
(1, 'Printer', 'Printing devices', '#28a745'),
(1, 'Network', 'Network equipment', '#ffc107'),
(1, 'Furniture', 'Office furniture', '#dc3545'),
(1, 'Vehicle', 'Company vehicles', '#6f42c1'),
(1, 'Equipment', 'Other equipment', '#fd7e14');

-- Create indexes for better performance
CREATE INDEX idx_assets_company ON assets(company_id);
CREATE INDEX idx_assets_institution ON assets(company_id, institution_name);
CREATE INDEX idx_assets_department ON assets(company_id, department);
CREATE INDEX idx_assets_status ON assets(company_id, status);
CREATE INDEX idx_assets_type ON assets(company_id, asset_type);
CREATE INDEX idx_assets_barcode ON assets(barcode);
CREATE INDEX idx_assets_qr_code ON assets(qr_code);
CREATE INDEX idx_users_company ON users(company_id);
CREATE INDEX idx_categories_company ON asset_categories(company_id); 