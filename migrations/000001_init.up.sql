CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    locale VARCHAR(10) DEFAULT 'en-US',
    is_active BOOLEAN DEFAULT true,
    username VARCHAR(50) UNIQUE,
    is_blocked BOOLEAN DEFAULT false,
    block_reason TEXT,
    blocked_at TIMESTAMP WITH TIME ZONE,
    blocked_by UUID,
    totp_secret VARCHAR(64),
    totp_enabled BOOLEAN DEFAULT false,
    totp_verified_at TIMESTAMP WITH TIME ZONE,
    avatar_url TEXT,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

ALTER TABLE users ADD CONSTRAINT fk_users_blocked_by FOREIGN KEY (blocked_by) REFERENCES users(id);

CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    is_blocked BOOLEAN NOT NULL DEFAULT false,
    block_reason TEXT,
    blocked_at TIMESTAMP,
    blocked_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at);
CREATE INDEX idx_tenants_is_blocked ON tenants(is_blocked) WHERE is_blocked = true;

CREATE TABLE IF NOT EXISTS tenant_members (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, user_id)
);

CREATE INDEX idx_tenant_members_tenant_id ON tenant_members(tenant_id);
CREATE INDEX idx_tenant_members_user_id ON tenant_members(user_id);

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_roles_name ON roles(name);

CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY,
    value VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

CREATE TABLE IF NOT EXISTS user_claims (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    claim_id UUID NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, claim_id)
);

CREATE INDEX idx_user_claims_user_id ON user_claims(user_id);
CREATE INDEX idx_user_claims_claim_id ON user_claims(claim_id);

CREATE TABLE IF NOT EXISTS role_claims (
    id UUID PRIMARY KEY,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    claim_id UUID NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, claim_id)
);

CREATE INDEX idx_role_claims_role_id ON role_claims(role_id);
CREATE INDEX idx_role_claims_claim_id ON role_claims(claim_id);

CREATE TABLE IF NOT EXISTS tenant_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_tenant_roles_tenant_id ON tenant_roles(tenant_id);

CREATE TABLE IF NOT EXISTS tenant_role_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_role_id UUID NOT NULL REFERENCES tenant_roles(id) ON DELETE CASCADE,
    claim_id UUID NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_role_id, claim_id)
);

CREATE INDEX idx_tenant_role_claims_tenant_role_id ON tenant_role_claims(tenant_role_id);
CREATE INDEX idx_tenant_role_claims_claim_id ON tenant_role_claims(claim_id);

CREATE TABLE IF NOT EXISTS user_tenant_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    tenant_role_id UUID NOT NULL REFERENCES tenant_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, tenant_id, tenant_role_id)
);

CREATE INDEX idx_user_tenant_roles_user_id ON user_tenant_roles(user_id);
CREATE INDEX idx_user_tenant_roles_tenant_id ON user_tenant_roles(tenant_id);
CREATE INDEX idx_user_tenant_roles_tenant_role_id ON user_tenant_roles(tenant_role_id);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_token_hash ON password_reset_tokens(token_hash);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);
CREATE INDEX idx_email_verification_tokens_expires_at ON email_verification_tokens(expires_at);

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    payload TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INTEGER NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMP,
    last_error_at TIMESTAMP,
    processed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbox_events_status_retry ON outbox_events(status, next_retry_at) WHERE status = 'pending';
CREATE INDEX idx_outbox_events_created_at ON outbox_events(created_at);

INSERT INTO roles (id, name, description, created_at, updated_at)
VALUES 
    ('00000000-0000-0000-0000-000000000001', 'admin', 'System administrator with full access', NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000002', 'user', 'Regular user with basic access', NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO claims (id, value, created_at)
VALUES 
    ('10000000-0000-0000-0000-000000000001', 'templates:view', NOW()),
    ('10000000-0000-0000-0000-000000000002', 'templates:create', NOW()),
    ('10000000-0000-0000-0000-000000000003', 'templates:update', NOW()),
    ('10000000-0000-0000-0000-000000000004', 'templates:delete', NOW()),
    ('10000000-0000-0000-0000-000000000011', 'members:view', NOW()),
    ('10000000-0000-0000-0000-000000000012', 'members:create', NOW()),
    ('10000000-0000-0000-0000-000000000013', 'members:update', NOW()),
    ('10000000-0000-0000-0000-000000000014', 'members:delete', NOW()),
    ('10000000-0000-0000-0000-000000000021', 'roles:view', NOW()),
    ('10000000-0000-0000-0000-000000000022', 'roles:create', NOW()),
    ('10000000-0000-0000-0000-000000000023', 'roles:update', NOW()),
    ('10000000-0000-0000-0000-000000000024', 'roles:delete', NOW()),
    ('10000000-0000-0000-0000-000000000031', 'tenant:view', NOW()),
    ('10000000-0000-0000-0000-000000000032', 'tenant:update', NOW()),
    ('10000000-0000-0000-0000-000000000033', 'tenant:delete', NOW()),
    ('10000000-0000-0000-0000-000000000034', 'tenant:manage', NOW()),
    ('10000000-0000-0000-0000-000000000041', 'branding:view', NOW()),
    ('10000000-0000-0000-0000-000000000042', 'branding:update', NOW()),
    ('10000000-0000-0000-0000-000000000051', 'analytics:view', NOW()),
    ('10000000-0000-0000-0000-000000000061', 'email:generate', NOW()),
    ('10000000-0000-0000-0000-000000000062', 'email:preview', NOW()),
    ('10000000-0000-0000-0000-000000000071', 'tenant:roles:view', NOW()),
    ('10000000-0000-0000-0000-000000000072', 'tenant:roles:manage', NOW()),
    ('10000000-0000-0000-0000-000000000081', 'tenant:members:view', NOW()),
    ('10000000-0000-0000-0000-000000000082', 'tenant:members:manage', NOW()),
    ('10000000-0000-0000-0000-000000000091', 'users:view', NOW()),
    ('10000000-0000-0000-0000-000000000092', 'users:manage', NOW()),
    ('10000000-0000-0000-0000-000000000093', 'claims:view', NOW()),
    ('10000000-0000-0000-0000-000000000094', 'claims:manage', NOW()),
    ('10000000-0000-0000-0000-000000000095', 'roles:view', NOW()),
    ('10000000-0000-0000-0000-000000000096', 'roles:manage', NOW()),
    ('10000000-0000-0000-0000-000000000097', 'support:manage', NOW()),
    ('10000000-0000-0000-0000-000000000101', 'push:view', NOW()),
    ('10000000-0000-0000-0000-000000000102', 'push:create', NOW()),
    ('10000000-0000-0000-0000-000000000103', 'push:update', NOW()),
    ('10000000-0000-0000-0000-000000000104', 'push:delete', NOW()),
    ('10000000-0000-0000-0000-000000000105', 'push:manage', NOW()),
    ('10000000-0000-0000-0000-000000000111', 'tenants:view', NOW()),
    ('10000000-0000-0000-0000-000000000112', 'tenants:manage', NOW()),
    ('10000000-0000-0000-0000-000000000121', 'tenant:secrets:view', NOW()),
    ('10000000-0000-0000-0000-000000000122', 'tenant:secrets:create', NOW()),
    ('10000000-0000-0000-0000-000000000123', 'tenant:secrets:edit', NOW()),
    ('10000000-0000-0000-0000-000000000124', 'tenant:secrets:delete', NOW()),
    ('10000000-0000-0000-0000-000000000125', 'tenant:secrets:manage', NOW())
ON CONFLICT (value) DO NOTHING;

INSERT INTO role_claims (id, role_id, claim_id, created_at)
SELECT 
    gen_random_uuid(),
    '00000000-0000-0000-0000-000000000001'::uuid,
    c.id,
    NOW()
FROM claims c
WHERE c.value IN (
    'users:manage',
    'claims:manage',
    'roles:manage',
    'support:manage',
    'push:manage',
    'tenants:manage'
)
ON CONFLICT (role_id, claim_id) DO NOTHING;
