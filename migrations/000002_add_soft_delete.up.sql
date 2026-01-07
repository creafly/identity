ALTER TABLE roles ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON roles(deleted_at);

ALTER TABLE tenant_roles ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_tenant_roles_deleted_at ON tenant_roles(deleted_at);
