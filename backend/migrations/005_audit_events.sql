CREATE TABLE IF NOT EXISTS audit_events (
  id VARCHAR(40) PRIMARY KEY,
  action VARCHAR(80) NOT NULL,
  actor_user_id VARCHAR(32) NOT NULL,
  actor_role VARCHAR(40) NOT NULL,
  request_id VARCHAR(40) NOT NULL,
  entity_type VARCHAR(60) NOT NULL,
  entity_id VARCHAR(40) NOT NULL,
  related_type VARCHAR(60) NULL,
  related_id VARCHAR(40) NULL,
  metadata JSON NOT NULL,
  created_at DATETIME NOT NULL,
  INDEX idx_audit_action_created (action, created_at),
  INDEX idx_audit_actor_created (actor_user_id, created_at),
  INDEX idx_audit_entity_created (entity_type, entity_id, created_at),
  INDEX idx_audit_related (related_type, related_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
