-- Drop the unique constraint
ALTER TABLE followers DROP CONSTRAINT IF EXISTS unique_follower_following;

-- Drop the indexes
DROP INDEX IF EXISTS idx_followers_follower_id;
DROP INDEX IF EXISTS idx_followers_following_id;

-- Drop the table
DROP TABLE IF EXISTS followers;