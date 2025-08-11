-- Canonical migration file for DB schema changes
-- MySQL-compatible (Amazon RDS MySQL). Run with the target DB selected (-D <db_name>).
-- Add additional ALTER/CREATE statements below as needed.

-- Example: Add companyId to all base tables if missing
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

CALL add_company_id_all_tables();
DROP PROCEDURE add_company_id_all_tables; 