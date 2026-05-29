SET @column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'uploads'
    AND COLUMN_NAME = 'confirmed_customer_id'
);

SET @alter_sql := IF(
  @column_exists = 0,
  'ALTER TABLE uploads ADD COLUMN confirmed_customer_id VARCHAR(32) NULL AFTER created_at',
  'SELECT 1'
);

PREPARE alter_uploads_confirmed_customer_id FROM @alter_sql;
EXECUTE alter_uploads_confirmed_customer_id;
DEALLOCATE PREPARE alter_uploads_confirmed_customer_id;

SET @column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'uploads'
    AND COLUMN_NAME = 'conversation_id'
);

SET @alter_sql := IF(
  @column_exists = 0,
  'ALTER TABLE uploads ADD COLUMN conversation_id VARCHAR(32) NULL AFTER confirmed_customer_id',
  'SELECT 1'
);

PREPARE alter_uploads_conversation_id FROM @alter_sql;
EXECUTE alter_uploads_conversation_id;
DEALLOCATE PREPARE alter_uploads_conversation_id;

SET @column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'uploads'
    AND COLUMN_NAME = 'agent_run_id'
);

SET @alter_sql := IF(
  @column_exists = 0,
  'ALTER TABLE uploads ADD COLUMN agent_run_id VARCHAR(32) NULL AFTER conversation_id',
  'SELECT 1'
);

PREPARE alter_uploads_agent_run_id FROM @alter_sql;
EXECUTE alter_uploads_agent_run_id;
DEALLOCATE PREPARE alter_uploads_agent_run_id;

SET @column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'uploads'
    AND COLUMN_NAME = 'followup_task_id'
);

SET @alter_sql := IF(
  @column_exists = 0,
  'ALTER TABLE uploads ADD COLUMN followup_task_id VARCHAR(32) NULL AFTER agent_run_id',
  'SELECT 1'
);

PREPARE alter_uploads_followup_task_id FROM @alter_sql;
EXECUTE alter_uploads_followup_task_id;
DEALLOCATE PREPARE alter_uploads_followup_task_id;
