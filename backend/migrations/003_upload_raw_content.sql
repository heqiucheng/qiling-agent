SET @column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'uploads'
    AND COLUMN_NAME = 'raw_content'
);

SET @alter_sql := IF(
  @column_exists = 0,
  'ALTER TABLE uploads ADD COLUMN raw_content MEDIUMTEXT NULL AFTER source_type',
  'SELECT 1'
);

PREPARE alter_uploads_raw_content FROM @alter_sql;
EXECUTE alter_uploads_raw_content;
DEALLOCATE PREPARE alter_uploads_raw_content;
