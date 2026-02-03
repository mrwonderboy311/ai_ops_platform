-- Drop discovered_hosts table
DROP INDEX IF EXISTS idx_discovered_hosts_unique;
DROP INDEX IF EXISTS idx_discovered_hosts_port;
DROP INDEX IF EXISTS idx_discovered_hosts_ip_address;
DROP INDEX IF EXISTS idx_discovered_hosts_scan_task_id;
DROP TABLE IF EXISTS discovered_hosts;

-- Drop scan_tasks table
DROP INDEX IF EXISTS idx_scan_tasks_created_at;
DROP INDEX IF EXISTS idx_scan_tasks_status;
DROP INDEX IF EXISTS idx_scan_tasks_user_id;
DROP TABLE IF EXISTS scan_tasks;
