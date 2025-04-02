-- Drop the unique constraint
ALTER TABLE blocked_users DROP CONSTRAINT IF EXISTS unique_user_id_blocked_user_id;

-- Drop the index
DROP INDEX IF EXISTS idx_blocked_users_user_id;

-- Drop the table
DROP TABLE IF EXISTS blocked_users;