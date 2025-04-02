CREATE TABLE error_logs (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    level VARCHAR(255) NOT NULL,
    message VARCHAR(255) NOT NULL,
    stack_trace TEXT,
    file_name VARCHAR(255),
    method_name VARCHAR(255),
    line_number INT,
    metadata JSON,
    ip_address VARCHAR(255),
    user_agent VARCHAR(255),
    occurred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    attempts INT DEFAULT 1,
    last_occurred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Remove the unique constraint from level
ALTER TABLE error_logs DROP INDEX uni_error_logs_level;

-- Add composite index for deduplication
CREATE INDEX idx_error_logs_dedupe ON error_logs (message(100), file_name, method_name, line_number, ip_address);