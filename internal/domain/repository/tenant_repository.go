package repository

import (
	"context"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *entity.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error)
	Update(ctx context.Context, tenant *entity.Tenant) error
	UpdateBlocked(ctx context.Context, id uuid.UUID, isBlocked bool, reason *string, blockedBy *uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.Tenant, error)
	AddMember(ctx context.Context, tenantID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error
	GetMembers(ctx context.Context, tenantID uuid.UUID) ([]*entity.User, error)
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*entity.Tenant, error)
	IsMember(ctx context.Context, tenantID, userID uuid.UUID) (bool, error)
}

type tenantRepository struct {
	db *sqlx.DB
}

func NewTenantRepository(db *sqlx.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *entity.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, display_name, slug, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.DisplayName, tenant.Slug, tenant.IsActive, tenant.CreatedAt, tenant.UpdatedAt,
	)
	return err
}

func (r *tenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	var tenant entity.Tenant
	query := `SELECT id, name, display_name, slug, is_active, is_blocked, block_reason, blocked_at, blocked_by, created_at, updated_at, deleted_at FROM tenants WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &tenant, query, id)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	var tenant entity.Tenant
	query := `SELECT id, name, display_name, slug, is_active, is_blocked, block_reason, blocked_at, blocked_by, created_at, updated_at, deleted_at FROM tenants WHERE slug = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &tenant, query, slug)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *entity.Tenant) error {
	query := `
		UPDATE tenants 
		SET name = $2, display_name = $3, slug = $4, is_active = $5, updated_at = $6
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.DisplayName, tenant.Slug, tenant.IsActive, tenant.UpdatedAt,
	)
	return err
}

func (r *tenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tenants SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *tenantRepository) UpdateBlocked(ctx context.Context, id uuid.UUID, isBlocked bool, reason *string, blockedBy *uuid.UUID) error {
	var query string
	var err error
	if isBlocked {
		query = `UPDATE tenants SET is_blocked = true, block_reason = $2, blocked_by = $3, blocked_at = NOW(), updated_at = NOW() WHERE id = $1`
		_, err = r.db.ExecContext(ctx, query, id, reason, blockedBy)
	} else {
		query = `UPDATE tenants SET is_blocked = false, block_reason = NULL, blocked_by = NULL, blocked_at = NULL, updated_at = NOW() WHERE id = $1`
		_, err = r.db.ExecContext(ctx, query, id)
	}
	return err
}

func (r *tenantRepository) List(ctx context.Context, offset, limit int) ([]*entity.Tenant, error) {
	var tenants []*entity.Tenant
	query := `SELECT id, name, display_name, slug, is_active, is_blocked, block_reason, blocked_at, blocked_by, created_at, updated_at, deleted_at FROM tenants WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &tenants, query, limit, offset)
	return tenants, err
}

func (r *tenantRepository) AddMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	query := `
		INSERT INTO tenant_members (id, tenant_id, user_id, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (tenant_id, user_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, utils.GenerateUUID(), tenantID, userID)
	return err
}

func (r *tenantRepository) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	query := `DELETE FROM tenant_members WHERE tenant_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, tenantID, userID)
	return err
}

func (r *tenantRepository) GetMembers(ctx context.Context, tenantID uuid.UUID) ([]*entity.User, error) {
	var users []*entity.User
	query := `
		SELECT u.* FROM users u
		JOIN tenant_members tm ON tm.user_id = u.id
		WHERE tm.tenant_id = $1 AND u.deleted_at IS NULL
	`
	err := r.db.SelectContext(ctx, &users, query, tenantID)
	return users, err
}

func (r *tenantRepository) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*entity.Tenant, error) {
	var tenants []*entity.Tenant
	query := `
		SELECT t.id, t.name, t.display_name, t.slug, t.is_active, t.is_blocked, t.block_reason, t.blocked_at, t.blocked_by, t.created_at, t.updated_at, t.deleted_at FROM tenants t
		JOIN tenant_members tm ON tm.tenant_id = t.id
		WHERE tm.user_id = $1 AND t.deleted_at IS NULL
	`
	err := r.db.SelectContext(ctx, &tenants, query, userID)
	return tenants, err
}

func (r *tenantRepository) IsMember(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM tenant_members tm
		JOIN tenants t ON t.id = tm.tenant_id
		WHERE tm.tenant_id = $1 AND tm.user_id = $2 AND t.deleted_at IS NULL
	`
	err := r.db.GetContext(ctx, &count, query, tenantID, userID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
