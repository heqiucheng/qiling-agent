CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(32) PRIMARY KEY,
  name VARCHAR(80) NOT NULL,
  role ENUM('sales', 'manager', 'admin') NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS customers (
  id VARCHAR(32) PRIMARY KEY,
  name VARCHAR(120) NOT NULL,
  source VARCHAR(80) NOT NULL,
  owner_id VARCHAR(32) NOT NULL,
  stage VARCHAR(40) NOT NULL,
  intent VARCHAR(20) NOT NULL,
  concerns JSON NOT NULL,
  tags JSON NOT NULL,
  profile_summary TEXT NOT NULL,
  last_contact_at DATETIME NOT NULL,
  pending_tasks INT NOT NULL DEFAULT 0,
  risk_flags JSON NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_customers_owner (owner_id),
  INDEX idx_customers_stage (stage),
  INDEX idx_customers_intent (intent),
  CONSTRAINT fk_customers_owner FOREIGN KEY (owner_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS conversation_messages (
  id VARCHAR(32) PRIMARY KEY,
  customer_id VARCHAR(32) NOT NULL,
  sender_type ENUM('customer', 'sales') NOT NULL,
  sender_name VARCHAR(120) NOT NULL,
  content TEXT NOT NULL,
  sent_at DATETIME NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_messages_customer_sent_at (customer_id, sent_at),
  CONSTRAINT fk_messages_customer FOREIGN KEY (customer_id) REFERENCES customers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS uploads (
  id VARCHAR(32) PRIMARY KEY,
  status VARCHAR(40) NOT NULL,
  source_type VARCHAR(40) NOT NULL,
  parsed_customer_name VARCHAR(120) NOT NULL,
  parsed_owner_name VARCHAR(120) NOT NULL,
  warnings JSON NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS followup_tasks (
  id VARCHAR(32) PRIMARY KEY,
  customer_id VARCHAR(32) NOT NULL,
  type VARCHAR(60) NOT NULL,
  status ENUM('pending', 'copied', 'skipped', 'marked_wrong') NOT NULL,
  generated_at DATETIME NOT NULL,
  recommendation JSON NOT NULL,
  feedback JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_tasks_customer (customer_id),
  INDEX idx_tasks_status (status),
  CONSTRAINT fk_tasks_customer FOREIGN KEY (customer_id) REFERENCES customers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS agent_runs (
  id VARCHAR(32) PRIMARY KEY,
  customer_id VARCHAR(32) NULL,
  task_type VARCHAR(80) NOT NULL,
  status VARCHAR(40) NOT NULL,
  model VARCHAR(80) NOT NULL,
  prompt_version VARCHAR(80) NOT NULL,
  input_summary TEXT NOT NULL,
  output JSON NULL,
  validation_errors JSON NOT NULL,
  risk_flags JSON NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at DATETIME NULL,
  INDEX idx_agent_runs_customer (customer_id),
  INDEX idx_agent_runs_status (status),
  CONSTRAINT fk_agent_runs_customer FOREIGN KEY (customer_id) REFERENCES customers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
