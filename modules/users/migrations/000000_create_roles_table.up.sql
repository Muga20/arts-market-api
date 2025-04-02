CREATE TABLE roles (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    role_name VARCHAR(100) NOT NULL,
    role_number INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true
);

-- Create index on role_name for faster lookups
CREATE INDEX idx_roles_role_name ON roles(role_name);