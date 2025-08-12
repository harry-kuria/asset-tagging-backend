-- Canonical migration: idempotent bootstrap for Asset Tagging schema
-- MySQL-compatible (Amazon RDS MySQL). Run with the target DB selected (-D asset_management).

-- 1) Create required tables if missing (minimal columns used by the app)
CREATE TABLE IF NOT EXISTS companies (
  id INT AUTO_INCREMENT PRIMARY KEY,
  company_name VARCHAR(255) NOT NULL,
  company_code VARCHAR(50) UNIQUE NOT NULL,
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

CREATE TABLE IF NOT EXISTS users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  company_id INT NOT NULL,
  username VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  role ENUM('admin','manager','user') DEFAULT 'user',
  is_active BOOLEAN DEFAULT TRUE,
  last_login TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY unique_username_per_company (company_id, username),
  UNIQUE KEY unique_email_per_company (company_id, email)
);

CREATE TABLE IF NOT EXISTS user_roles (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  company_id INT NOT NULL,
  role VARCHAR(100) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS asset_categories (
  id INT AUTO_INCREMENT PRIMARY KEY,
  company_id INT NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  color VARCHAR(7) DEFAULT '#007bff',
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 2) Helper: add a column if it does not already exist
DELIMITER $$
DROP PROCEDURE IF EXISTS add_col_if_missing $$
CREATE PROCEDURE add_col_if_missing(
  IN p_schema VARCHAR(64),
  IN p_table  VARCHAR(64),
  IN p_column VARCHAR(64),
  IN p_definition TEXT
)
BEGIN
  DECLARE col_count INT;
  SELECT COUNT(*) INTO col_count
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = p_schema AND TABLE_NAME = p_table AND COLUMN_NAME = p_column;
  IF col_count = 0 THEN
    SET @ddl = CONCAT('ALTER TABLE `', p_schema, '`.`', p_table, '` ADD COLUMN ', p_definition);
    PREPARE s FROM @ddl; EXECUTE s; DEALLOCATE PREPARE s;
  END IF;
END $$
DELIMITER ;

-- 3) Ensure required columns exist (safe to run multiple times)
CALL add_col_if_missing(DATABASE(), 'users', 'company_id', 'INT NOT NULL');
CALL add_col_if_missing(DATABASE(), 'users', 'email', 'VARCHAR(255) NOT NULL');
CALL add_col_if_missing(DATABASE(), 'users', 'password_hash', 'VARCHAR(255) NOT NULL');
CALL add_col_if_missing(DATABASE(), 'users', 'first_name', 'VARCHAR(100) NULL');
CALL add_col_if_missing(DATABASE(), 'users', 'last_name', 'VARCHAR(100) NULL');
CALL add_col_if_missing(DATABASE(), 'users', 'role', "ENUM('admin','manager','user') DEFAULT 'user'");
CALL add_col_if_missing(DATABASE(), 'users', 'is_active', 'BOOLEAN DEFAULT TRUE');
CALL add_col_if_missing(DATABASE(), 'users', 'last_login', 'TIMESTAMP NULL');
CALL add_col_if_missing(DATABASE(), 'users', 'created_at', 'TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP');
CALL add_col_if_missing(DATABASE(), 'users', 'updated_at', 'TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP');

CALL add_col_if_missing(DATABASE(), 'user_roles', 'company_id', 'INT NOT NULL');
CALL add_col_if_missing(DATABASE(), 'user_roles', 'created_at', 'TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP');

CALL add_col_if_missing(DATABASE(), 'asset_categories', 'company_id', 'INT NOT NULL');
CALL add_col_if_missing(DATABASE(), 'asset_categories', 'description', 'TEXT NULL');
CALL add_col_if_missing(DATABASE(), 'asset_categories', 'color', 'VARCHAR(7) NULL');
CALL add_col_if_missing(DATABASE(), 'asset_categories', 'is_active', 'BOOLEAN DEFAULT TRUE');
CALL add_col_if_missing(DATABASE(), 'asset_categories', 'created_at', 'TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP');
CALL add_col_if_missing(DATABASE(), 'asset_categories', 'updated_at', 'TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP');

-- 4) Optional legacy support: ensure companyId (camelCase) exists everywhere
DELIMITER $$
DROP PROCEDURE IF EXISTS add_company_id_all_tables $$
CREATE PROCEDURE add_company_id_all_tables()
BEGIN
  DECLARE done INT DEFAULT FALSE;
  DECLARE v_table_name VARCHAR(255);

  DECLARE cur CURSOR FOR
    SELECT t.TABLE_NAME
    FROM INFORMATION_SCHEMA.TABLES t
    WHERE t.TABLE_SCHEMA = DATABASE()
      AND t.TABLE_TYPE = 'BASE TABLE'
      AND NOT EXISTS (
        SELECT 1
        FROM INFORMATION_SCHEMA.COLUMNS c
        WHERE c.TABLE_SCHEMA = t.TABLE_SCHEMA
          AND c.TABLE_NAME = t.TABLE_NAME
          AND c.COLUMN_NAME = 'companyId'
      );

  DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;

  OPEN cur;
  read_loop: LOOP
    FETCH cur INTO v_table_name;
    IF done THEN
      LEAVE read_loop;
    END IF;

    SET @stmt = CONCAT('ALTER TABLE `', REPLACE(v_table_name, '`', '``'),
                       '` ADD COLUMN `companyId` VARCHAR(64) NULL');
    PREPARE s FROM @stmt;
    EXECUTE s;
    DEALLOCATE PREPARE s;
  END LOOP;
  CLOSE cur;
END $$
DELIMITER ;

-- Run legacy companyId addition (no-op for tables that already have it)
CALL add_company_id_all_tables();

-- Cleanup helpers
DROP PROCEDURE add_company_id_all_tables;
DROP PROCEDURE add_col_if_missing; 