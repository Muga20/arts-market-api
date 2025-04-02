CREATE TABLE user_locations (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36),
    country VARCHAR(100),
    state VARCHAR(100),
    state_name VARCHAR(100),
    continent VARCHAR(100),
    city VARCHAR(100),
    zip VARCHAR(20),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_user_locations_user_id ON user_locations(user_id);