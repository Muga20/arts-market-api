CREATE TABLE notifications (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()), -- MySQL: UUID(), PostgreSQL: uuid_generate_v4()
    user_id CHAR(36) NOT NULL,               -- Recipient of the notification
    sender_id CHAR(36),                      -- Who triggered it (NULL for system)
    notification_type VARCHAR(50) NOT NULL,  -- e.g., "follow", "email_update", "post_liked"
    message TEXT NOT NULL,                   -- Human-readable message
    is_read BOOLEAN NOT NULL DEFAULT false,  -- Read status
    is_system_generated BOOLEAN NOT NULL DEFAULT false, -- System vs. user-triggered
    entity_type VARCHAR(50),                 -- e.g., "user", "post", "email" (what’s affected)
    entity_id CHAR(36),                      -- ID of the affected entity (e.g., post ID)
    priority INT NOT NULL DEFAULT 0,         -- 0=low, 1=medium, 2=high (for sorting/filtering)
    metadata JSON,                           -- Extra data (e.g., old/new email values)
    expires_at TIMESTAMP,                    -- Optional expiration for transient notifications
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Create indexes for performance
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_sender_id ON notifications(sender_id);
CREATE INDEX idx_notifications_entity_type_entity_id ON notifications(entity_type, entity_id);