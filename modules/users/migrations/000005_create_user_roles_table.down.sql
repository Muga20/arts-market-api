-- Drop the unique constraint
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS unique_user_role;

-- Drop the index
DROP INDEX IF EXISTS idx_user_roles_user_id_role_id;

-- Drop the table
DROP TABLE IF EXISTS user_roles;