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
-- Check if company_id column already exists
SET @company_id_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
                         WHERE TABLE_SCHEMA = DATABASE() 
                         AND TABLE_NAME = 'assets' 
                         AND COLUMN_NAME = 'company_id');

-- Add company_id column if it doesn't exist
SET @sql = IF(@company_id_exists = 0, 
    'ALTER TABLE assets ADD COLUMN company_id INT NULL', 
    'SELECT "company_id column already exists" as message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Check if companyId column exists before trying to update and drop it
SET @companyId_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
                        WHERE TABLE_SCHEMA = DATABASE() 
                        AND TABLE_NAME = 'assets' 
                        AND COLUMN_NAME = 'companyId');

-- Update existing assets to use company_id = 1 (default company) if companyId column exists
SET @sql = IF(@companyId_exists > 0, 
    'UPDATE assets SET company_id = 1 WHERE companyId = "1" OR companyId IS NULL OR company_id IS NULL', 
    'UPDATE assets SET company_id = 1 WHERE company_id IS NULL');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Drop the old companyId column if it exists
SET @sql = IF(@companyId_exists > 0, 
    'ALTER TABLE assets DROP COLUMN companyId', 
    'SELECT "companyId column does not exist" as message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Check if foreign key constraint already exists
SET @fk_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
                  WHERE TABLE_SCHEMA = DATABASE() 
                  AND TABLE_NAME = 'assets' 
                  AND CONSTRAINT_NAME = 'fk_assets_company');

-- Add foreign key constraint if it doesn't exist
SET @sql = IF(@fk_exists = 0, 
    'ALTER TABLE assets ADD CONSTRAINT fk_assets_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE', 
    'SELECT "foreign key constraint already exists" as message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Create indexes for better performance (with error handling)
-- Check if indexes exist before creating them
SET @idx_company_id_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
                             WHERE TABLE_SCHEMA = DATABASE() 
                             AND TABLE_NAME = 'assets' 
                             AND INDEX_NAME = 'idx_assets_company_id');

SET @idx_status_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
                         WHERE TABLE_SCHEMA = DATABASE() 
                         AND TABLE_NAME = 'assets' 
                         AND INDEX_NAME = 'idx_assets_status');

SET @idx_asset_type_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
                             WHERE TABLE_SCHEMA = DATABASE() 
                             AND TABLE_NAME = 'assets' 
                             AND INDEX_NAME = 'idx_assets_asset_type');

SET @idx_users_company_id_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
                                   WHERE TABLE_SCHEMA = DATABASE() 
                                   AND TABLE_NAME = 'users' 
                                   AND INDEX_NAME = 'idx_users_company_id');

SET @idx_users_username_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
                                 WHERE TABLE_SCHEMA = DATABASE() 
                                 AND TABLE_NAME = 'users' 
                                 AND INDEX_NAME = 'idx_users_username');

-- Create indexes only if they don't exist
SET @sql = IF(@idx_company_id_exists = 0, 
    'CREATE INDEX idx_assets_company_id ON assets(company_id)', 
    'SELECT "idx_assets_company_id already exists" as message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = IF(@idx_status_exists = 0, 
    'CREATE INDEX idx_assets_status ON assets(status)', 
    'SELECT "idx_assets_status already exists" as message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = IF(@idx_asset_type_exists = 0, 
    'CREATE INDEX idx_assets_asset_type ON assets(asset_type)', 
    'SELECT "idx_assets_asset_type already exists" as message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = IF(@idx_users_company_id_exists = 0, 
    'CREATE INDEX idx_users_company_id ON users(company_id)', 
    'SELECT "idx_users_company_id already exists" as message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = IF(@idx_users_username_exists = 0, 
    'CREATE INDEX idx_users_username ON users(username)', 
    'SELECT "idx_users_username already exists" as message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt; 