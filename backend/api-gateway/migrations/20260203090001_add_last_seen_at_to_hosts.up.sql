-- Add last_seen_at column to hosts table
ALTER TABLE hosts ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_hosts_last_seen_at ON hosts(last_seen_at);
