INSERT INTO claims (id, value, created_at)
VALUES 
    ('10000000-0000-0000-0000-000000000131', 'tenant:mcp:view', NOW()),
    ('10000000-0000-0000-0000-000000000132', 'tenant:mcp:create', NOW()),
    ('10000000-0000-0000-0000-000000000133', 'tenant:mcp:delete', NOW()),
    ('10000000-0000-0000-0000-000000000134', 'tenant:mcp:manage', NOW())
ON CONFLICT (value) DO NOTHING;
