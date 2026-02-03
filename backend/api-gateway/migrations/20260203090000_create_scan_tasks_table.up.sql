-- Create scan_tasks table for SSH scanning
CREATE TABLE IF NOT EXISTS scan_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_range VARCHAR(100) NOT NULL,
    ports INTEGER[] NOT NULL DEFAULT '{22}',
    timeout_seconds INTEGER DEFAULT 5,
    status VARCHAR(50) DEFAULT 'running',
    estimated_hosts INTEGER DEFAULT 0,
    discovered_hosts INTEGER DEFAULT 0,
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create index on user_id for faster user-specific queries
CREATE INDEX idx_scan_tasks_user_id ON scan_tasks(user_id);
CREATE INDEX idx_scan_tasks_status ON scan_tasks(status);
CREATE INDEX idx_scan_tasks_created_at ON scan_tasks(created_at DESC);

-- Create discovered_hosts table to store scan results
CREATE TABLE IF NOT EXISTS discovered_hosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_task_id UUID NOT NULL REFERENCES scan_tasks(id) ON DELETE CASCADE,
    ip_address INET NOT NULL,
    port INTEGER NOT NULL,
    hostname VARCHAR(255),
    os_type VARCHAR(100),
    os_version VARCHAR(100),
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create index on scan_task_id for faster task-specific queries
CREATE INDEX idx_discovered_hosts_scan_task_id ON discovered_hosts(scan_task_id);
CREATE INDEX idx_discovered_hosts_ip_address ON discovered_hosts(ip_address);
CREATE INDEX idx_discovered_hosts_port ON discovered_hosts(port);

-- Create unique constraint to avoid duplicate entries per scan
CREATE UNIQUE INDEX idx_discovered_hosts_unique ON discovered_hosts(scan_task_id, ip_address, port);

-- Add comments for documentation
COMMENT ON TABLE scan_tasks IS 'Tracks SSH network scanning tasks';
COMMENT ON TABLE discovered_hosts IS 'Stores results from SSH scan tasks';
COMMENT ON COLUMN scan_tasks.status IS 'running, completed, failed, or cancelled';
COMMENT ON COLUMN discovered_hosts.status IS 'success, open, timeout, or error';
