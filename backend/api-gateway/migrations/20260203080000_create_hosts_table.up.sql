-- Create hosts table
CREATE TABLE hosts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  hostname VARCHAR(255) NOT NULL,
  ip_address INET NOT NULL,
  port INTEGER DEFAULT 22,
  status VARCHAR(50) DEFAULT 'pending',
  os_type VARCHAR(100),
  os_version VARCHAR(100),
  cpu_cores INTEGER,
  memory_gb INTEGER,
  disk_gb BIGINT,
  labels JSONB DEFAULT '{}',
  tags TEXT[] DEFAULT '{}',
  cluster_id UUID,
  registered_by UUID REFERENCES users(id) ON DELETE SET NULL,
  approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
  approved_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  CONSTRAINT unique_host UNIQUE(ip_address, port)
);

-- Create indexes for hosts table
CREATE INDEX idx_hosts_status ON hosts(status);
CREATE INDEX idx_hosts_hostname ON hosts(hostname);
CREATE INDEX idx_hosts_labels ON hosts USING GIN(labels);
CREATE INDEX idx_hosts_registered_by ON hosts(registered_by);
CREATE INDEX idx_hosts_approved_by ON hosts(approved_by);

-- Add comment to table
COMMENT ON TABLE hosts IS 'Host inventory table for managed servers';
COMMENT ON COLUMN hosts.status IS 'Host status: pending, approved, rejected, offline, online';
COMMENT ON COLUMN hosts.labels IS 'Flexible key-value labels for grouping and filtering';
COMMENT ON COLUMN hosts.tags IS 'Text tags for categorization';
