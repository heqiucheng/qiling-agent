CREATE TABLE IF NOT EXISTS report_export_tasks (
  id VARCHAR(64) PRIMARY KEY,
  report_id VARCHAR(64) NOT NULL,
  export_format VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  owner_id VARCHAR(64) NOT NULL,
  owner_role VARCHAR(32) NOT NULL,
  filename VARCHAR(255) NOT NULL,
  content_type VARCHAR(128) NOT NULL,
  size_bytes INT NOT NULL DEFAULT 0,
  error_message TEXT NULL,
  created_at DATETIME NOT NULL,
  completed_at DATETIME NULL,
  INDEX idx_report_export_tasks_owner_time (owner_id, owner_role, created_at),
  INDEX idx_report_export_tasks_report_format (report_id, export_format)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
