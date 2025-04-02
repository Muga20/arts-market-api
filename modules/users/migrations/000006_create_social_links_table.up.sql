CREATE TABLE
    social_links (
        id CHAR(36) PRIMARY KEY DEFAULT (UUID ()),
        user_id CHAR(36) NOT NULL,
        platform VARCHAR(50) NOT NULL,
        link VARCHAR(255) NOT NULL,
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
    );

-- Create index on user_id for faster lookups
CREATE INDEX idx_social_links_user_id ON social_links (user_id);