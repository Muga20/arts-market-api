CREATE TABLE user_security (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_email_verified_at TIMESTAMP,
    login_attempts INT NOT NULL DEFAULT 0,
    last_login_attempt_at TIMESTAMP,
    is_locked BOOLEAN NOT NULL DEFAULT false,
    locked_at TIMESTAMP,
    unlock_token VARCHAR(255),
    unlock_token_expires_at TIMESTAMP,
    last_successful_login_at TIMESTAMP,
    password_reset_token VARCHAR(255),
    password_reset_expires_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_user_security_user_id ON user_security(user_id);