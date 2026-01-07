DROP INDEX IF EXISTS idx_roles_deleted_at;
ALTER TABLE roles DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_tenant_roles_deleted_at;
ALTER TABLE tenant_roles DROP COLUMN IF EXISTS deleted_at;
