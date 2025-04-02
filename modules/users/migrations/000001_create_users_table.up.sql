CREATE TABLE users (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50),
    username VARCHAR(50),
    status VARCHAR(50),
    auth_type VARCHAR(50),
    is_active BOOLEAN NOT NULL DEFAULT true,
    deleted_at TIMESTAMP
);

-- Create index on email for faster lookups
CREATE INDEX idx_users_email ON users(email);

-- Add a unique constraint on email
ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email);