DELETE FROM claims 
WHERE value IN (
    'tenant:mcp:view',
    'tenant:mcp:create',
    'tenant:mcp:delete',
    'tenant:mcp:manage'
);
