-- Drop the indexes
DROP INDEX IF EXISTS idx_notifications_user_id;
DROP INDEX IF EXISTS idx_notifications_sender_id;
DROP INDEX IF EXISTS idx_notifications_entity_type_entity_id;

-- Drop the table
DROP TABLE IF EXISTS notifications;