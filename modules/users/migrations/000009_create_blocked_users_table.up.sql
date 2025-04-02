CREATE TABLE blocked_users (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL, -- Who blocked
    blocked_user_id CHAR(36) NOT NULL, -- Who is blocked
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (blocked_user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (user_id, blocked_user_id)
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_blocked_users_user_id ON blocked_users(user_id);