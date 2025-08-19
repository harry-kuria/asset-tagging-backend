-- Migration: Create companies table and update assets to use auto-incrementing company IDs
-- Date: 2025-01-27

-- Create companies table if it doesn't exist
CREATE TABLE IF NOT EXISTS companies (
    id INT AUTO_INCREMENT PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    company_code VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255),
    subscription_plan VARCHAR(50) DEFAULT 'trial',
    is_active BOOLEAN DEFAULT TRUE,
    trial_ends_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Create users table if it doesn't exist
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    company_id INT NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
);

-- Create a default company if none exists
INSERT IGNORE INTO companies (id, company_name, company_code, email, subscription_plan, is_active, trial_ends_at) 
VALUES (1, 'Default Company', 'DEFAULT', 'admin@default.com', 'trial', TRUE, DATE_ADD(NOW(), INTERVAL 30 DAY));

-- Update assets table to use integer company_id instead of varchar companyId
-- First, add a new column
ALTER TABLE assets ADD COLUMN company_id INT NULL;

-- Update existing assets to use company_id = 1 (default company)
UPDATE assets SET company_id = 1 WHERE companyId = '1' OR companyId IS NULL;

-- Drop the old companyId column
ALTER TABLE assets DROP COLUMN companyId;

-- Add foreign key constraint
ALTER TABLE assets ADD CONSTRAINT fk_assets_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE;

-- Create indexes for better performance
CREATE INDEX idx_assets_company_id ON assets(company_id);
CREATE INDEX idx_assets_status ON assets(status);
CREATE INDEX idx_assets_asset_type ON assets(asset_type);
CREATE INDEX idx_users_company_id ON users(company_id);
CREATE INDEX idx_users_username ON users(username); 