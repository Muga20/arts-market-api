-- Drop the unique constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS unique_email;

-- Drop the index
DROP INDEX IF EXISTS idx_users_email;

-- Drop the table
DROP TABLE IF EXISTS users;