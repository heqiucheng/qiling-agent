CREATE TABLE IF NOT EXISTS customer_memory_facts (
  id VARCHAR(40) PRIMARY KEY,
  customer_id VARCHAR(32) NOT NULL,
  category VARCHAR(60) NOT NULL,
  fact_key VARCHAR(120) NOT NULL,
  fact_value TEXT NOT NULL,
  confidence DECIMAL(4,3) NOT NULL,
  source_type VARCHAR(60) NOT NULL,
  source_id VARCHAR(40) NOT NULL,
  status ENUM('active', 'superseded', 'rejected') NOT NULL DEFAULT 'active',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uniq_memory_customer_fact (customer_id, category, fact_key),
  INDEX idx_memory_customer_status_updated (customer_id, status, updated_at),
  INDEX idx_memory_source (source_type, source_id),
  CONSTRAINT fk_memory_customer FOREIGN KEY (customer_id) REFERENCES customers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
