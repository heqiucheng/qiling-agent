CREATE TABLE IF NOT EXISTS saved_reports (
  id VARCHAR(64) PRIMARY KEY,
  report_type VARCHAR(64) NOT NULL,
  title VARCHAR(255) NOT NULL,
  range_label VARCHAR(64) NOT NULL,
  summary TEXT NOT NULL,
  owner_id VARCHAR(64) NOT NULL,
  owner_role VARCHAR(32) NOT NULL,
  metrics_count INT NOT NULL DEFAULT 0,
  sections_count INT NOT NULL DEFAULT 0,
  action_items_count INT NOT NULL DEFAULT 0,
  report_json JSON NOT NULL,
  markdown MEDIUMTEXT NOT NULL,
  generated_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  INDEX idx_saved_reports_owner_time (owner_id, owner_role, generated_at),
  INDEX idx_saved_reports_type_time (report_type, generated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
