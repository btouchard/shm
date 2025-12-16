CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL UNIQUE,
    public_key BYTEA NOT NULL,
    app_name VARCHAR(100) DEFAULT 'unknown_app',
    app_version VARCHAR(50),          -- Ex: "1.0.2", "v2.3.0-beta"
    deployment_mode VARCHAR(20),      -- Ex: "docker", "binary", "k8s"
    environment VARCHAR(20),          -- Ex: "production", "staging"
    os_arch VARCHAR(50),              -- Ex: "linux/amd64"
    status VARCHAR(20) DEFAULT 'pending', -- pending, active, revoked
    last_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE snapshots (
    id BIGSERIAL PRIMARY KEY,
    instance_id UUID NOT NULL REFERENCES instances(instance_id) ON DELETE CASCADE,
    snapshot_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    data JSONB NOT NULL
);

CREATE INDEX idx_instances_app_name ON instances(app_name);
CREATE INDEX idx_instances_last_seen ON instances(last_seen_at);
