CREATE TABLE user_roles (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    role_id CHAR(36) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- Create composite index on user_id and role_id for faster lookups
CREATE INDEX idx_user_roles_user_id_role_id ON user_roles(user_id, role_id);

-- Ensure a user can't have the same role assigned multiple times
ALTER TABLE user_roles ADD CONSTRAINT unique_user_role UNIQUE (user_id, role_id);