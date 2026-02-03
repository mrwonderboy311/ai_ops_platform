-- Remove last_seen_at column from hosts table
DROP INDEX IF EXISTS idx_hosts_last_seen_at;
ALTER TABLE hosts DROP COLUMN IF EXISTS last_seen_at;
