CREATE TABLE user_privacy_settings (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    show_email BOOLEAN NOT NULL DEFAULT false,
    show_phone BOOLEAN NOT NULL DEFAULT false,
    show_location BOOLEAN NOT NULL DEFAULT false,
    allow_follow_requests BOOLEAN NOT NULL DEFAULT true,
    allow_messages_from VARCHAR(20) NOT NULL DEFAULT 'everyone', -- 'everyone', 'following', 'none'
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_user_privacy_settings_user_id ON user_privacy_settings(user_id);