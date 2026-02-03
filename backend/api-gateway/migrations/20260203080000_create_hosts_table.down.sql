-- Drop indexes first
DROP INDEX IF EXISTS idx_hosts_approved_by;
DROP INDEX IF EXISTS idx_hosts_registered_by;
DROP INDEX IF EXISTS idx_hosts_labels;
DROP INDEX IF EXISTS idx_hosts_hostname;
DROP INDEX IF EXISTS idx_hosts_status;

-- Drop hosts table
DROP TABLE IF EXISTS hosts;
