package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/creafly/identity/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

type TenantRoleRepository interface {
	Create(ctx context.Context, role *entity.TenantRole) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.TenantRole, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*entity.TenantRole, error)
	Update(ctx context.Context, role *entity.TenantRole) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantRole, error)
	GetDefaultRoles(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantRole, error)

	AddClaim(ctx context.Context, tenantRoleID, claimID uuid.UUID) error
	RemoveClaim(ctx context.Context, tenantRoleID, claimID uuid.UUID) error
	GetRoleClaims(ctx context.Context, tenantRoleID uuid.UUID) ([]*entity.Claim, error)
	GetTenantAvailableClaims(ctx context.Context, tenantID uuid.UUID) ([]*entity.Claim, error)
	BatchUpdateClaims(ctx context.Context, tenantRoleID uuid.UUID, assignClaimIDs, removeClaimIDs []uuid.UUID) error

	AssignToUser(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error
	RemoveFromUser(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error
	RemoveAllFromUser(ctx context.Context, userID, tenantID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.TenantRole, error)
	GetUserClaims(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.Claim, error)

	CreateDefaultRoles(ctx context.Context, tenantID uuid.UUID) error
}

type tenantRoleRepository struct {
	db *sqlx.DB
}

func NewTenantRoleRepository(db *sqlx.DB) TenantRoleRepository {
	return &tenantRoleRepository{db: db}
}

func (r *tenantRoleRepository) Create(ctx context.Context, role *entity.TenantRole) error {
	query := `
		INSERT INTO tenant_roles (id, tenant_id, name, description, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		role.ID, role.TenantID, role.Name, role.Description, role.IsDefault, role.CreatedAt, role.UpdatedAt,
	)
	return err
}

func (r *tenantRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.TenantRole, error) {
	var role entity.TenantRole
	query := `SELECT id, tenant_id, name, description, is_default, created_at, updated_at FROM tenant_roles WHERE id = $1`
	err := r.db.GetContext(ctx, &role, query, id)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *tenantRoleRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*entity.TenantRole, error) {
	var role entity.TenantRole
	query := `SELECT id, tenant_id, name, description, is_default, created_at, updated_at FROM tenant_roles WHERE tenant_id = $1 AND name = $2`
	err := r.db.GetContext(ctx, &role, query, tenantID, name)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *tenantRoleRepository) Update(ctx context.Context, role *entity.TenantRole) error {
	query := `
		UPDATE tenant_roles 
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, role.ID, role.Name, role.Description, role.UpdatedAt)
	return err
}

func (r *tenantRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenant_roles WHERE id = $1 AND is_default = false`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *tenantRoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantRole, error) {
	var roles []*entity.TenantRole
	query := `SELECT id, tenant_id, name, description, is_default, created_at, updated_at FROM tenant_roles WHERE tenant_id = $1 ORDER BY is_default DESC, name`
	err := r.db.SelectContext(ctx, &roles, query, tenantID)
	return roles, err
}

func (r *tenantRoleRepository) GetDefaultRoles(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantRole, error) {
	var roles []*entity.TenantRole
	query := `SELECT id, tenant_id, name, description, is_default, created_at, updated_at FROM tenant_roles WHERE tenant_id = $1 AND is_default = true ORDER BY name`
	err := r.db.SelectContext(ctx, &roles, query, tenantID)
	return roles, err
}

func (r *tenantRoleRepository) AddClaim(ctx context.Context, tenantRoleID, claimID uuid.UUID) error {
	query := `
		INSERT INTO tenant_role_claims (id, tenant_role_id, claim_id, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (tenant_role_id, claim_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, uuid.New(), tenantRoleID, claimID)
	return err
}

func (r *tenantRoleRepository) RemoveClaim(ctx context.Context, tenantRoleID, claimID uuid.UUID) error {
	query := `DELETE FROM tenant_role_claims WHERE tenant_role_id = $1 AND claim_id = $2`
	_, err := r.db.ExecContext(ctx, query, tenantRoleID, claimID)
	return err
}

func (r *tenantRoleRepository) GetRoleClaims(ctx context.Context, tenantRoleID uuid.UUID) ([]*entity.Claim, error) {
	var claims []*entity.Claim
	query := `
		SELECT c.id, c.value, c.created_at
		FROM claims c
		JOIN tenant_role_claims trc ON trc.claim_id = c.id
		WHERE trc.tenant_role_id = $1
		ORDER BY c.value
	`
	err := r.db.SelectContext(ctx, &claims, query, tenantRoleID)
	return claims, err
}

func (r *tenantRoleRepository) AssignToUser(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error {
	query := `
		INSERT INTO user_tenant_roles (id, user_id, tenant_id, tenant_role_id, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, tenant_id, tenant_role_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, uuid.New(), userID, tenantID, tenantRoleID)
	return err
}

func (r *tenantRoleRepository) RemoveFromUser(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error {
	query := `DELETE FROM user_tenant_roles WHERE user_id = $1 AND tenant_id = $2 AND tenant_role_id = $3`
	_, err := r.db.ExecContext(ctx, query, userID, tenantID, tenantRoleID)
	return err
}

func (r *tenantRoleRepository) RemoveAllFromUser(ctx context.Context, userID, tenantID uuid.UUID) error {
	query := `DELETE FROM user_tenant_roles WHERE user_id = $1 AND tenant_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, tenantID)
	return err
}

func (r *tenantRoleRepository) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.TenantRole, error) {
	var roles []*entity.TenantRole
	query := `
		SELECT tr.id, tr.tenant_id, tr.name, tr.description, tr.is_default, tr.created_at, tr.updated_at 
		FROM tenant_roles tr
		JOIN user_tenant_roles utr ON utr.tenant_role_id = tr.id
		WHERE utr.user_id = $1 AND utr.tenant_id = $2
		ORDER BY tr.name
	`
	err := r.db.SelectContext(ctx, &roles, query, userID, tenantID)
	return roles, err
}

func (r *tenantRoleRepository) GetUserClaims(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.Claim, error) {
	var claims []*entity.Claim
	query := `
		SELECT DISTINCT c.id, c.value, c.created_at
		FROM claims c
		JOIN tenant_role_claims trc ON trc.claim_id = c.id
		JOIN user_tenant_roles utr ON utr.tenant_role_id = trc.tenant_role_id
		WHERE utr.user_id = $1 AND utr.tenant_id = $2
		ORDER BY c.value
	`
	err := r.db.SelectContext(ctx, &claims, query, userID, tenantID)
	return claims, err
}

func (r *tenantRoleRepository) GetTenantAvailableClaims(ctx context.Context, tenantID uuid.UUID) ([]*entity.Claim, error) {
	var claims []*entity.Claim
	query := `
		SELECT DISTINCT c.id, c.value, c.created_at
		FROM claims c
		JOIN tenant_role_claims trc ON trc.claim_id = c.id
		JOIN tenant_roles tr ON tr.id = trc.tenant_role_id
		WHERE tr.tenant_id = $1
		ORDER BY c.value
	`
	err := r.db.SelectContext(ctx, &claims, query, tenantID)
	return claims, err
}

func (r *tenantRoleRepository) CreateDefaultRoles(ctx context.Context, tenantID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var claims []*entity.Claim
	err = tx.SelectContext(ctx, &claims, `SELECT id, value, created_at FROM claims ORDER BY value`)
	if err != nil {
		return err
	}

	claimMap := make(map[string]uuid.UUID)
	for _, c := range claims {
		claimMap[c.Value] = c.ID
	}

	defaultRoles := []struct {
		Name        string
		Description string
		Claims      []string
	}{
		{
			Name:        "owner",
			Description: "Tenant owner with full access",
			Claims:      []string{},
		},
		{
			Name:        "admin",
			Description: "Tenant administrator",
			Claims: []string{
				"templates:view", "templates:create", "templates:update", "templates:delete",
				"members:view", "members:create", "members:update", "members:delete",
				"roles:view", "roles:create", "roles:update", "roles:delete",
				"tenant:view", "tenant:update", "tenant:manage",
				"tenant:roles:view", "tenant:roles:manage",
				"tenant:members:view", "tenant:members:manage",
				"branding:view", "branding:update",
				"analytics:view",
				"email:generate", "email:preview",
			},
		},
		{
			Name:        "member",
			Description: "Regular tenant member",
			Claims: []string{
				"templates:view", "templates:create", "templates:update",
				"members:view",
				"tenant:view",
				"tenant:roles:view",
				"tenant:members:view",
				"branding:view",
				"analytics:view",
				"email:generate", "email:preview",
			},
		},
	}

	for _, roleDef := range defaultRoles {
		roleID := uuid.New()

		_, err = tx.ExecContext(ctx, `
			INSERT INTO tenant_roles (id, tenant_id, name, description, is_default, created_at, updated_at)
			VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		`, roleID, tenantID, roleDef.Name, roleDef.Description)
		if err != nil {
			return err
		}

		if roleDef.Name == "owner" {
			for _, c := range claims {
				_, err = tx.ExecContext(ctx, `
					INSERT INTO tenant_role_claims (id, tenant_role_id, claim_id, created_at)
					VALUES ($1, $2, $3, NOW())
				`, uuid.New(), roleID, c.ID)
				if err != nil {
					return err
				}
			}
		} else {
			for _, claimValue := range roleDef.Claims {
				if claimID, ok := claimMap[claimValue]; ok {
					_, err = tx.ExecContext(ctx, `
						INSERT INTO tenant_role_claims (id, tenant_role_id, claim_id, created_at)
						VALUES ($1, $2, $3, NOW())
					`, uuid.New(), roleID, claimID)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return tx.Commit()
}

func (r *tenantRoleRepository) BatchUpdateClaims(ctx context.Context, tenantRoleID uuid.UUID, assignClaimIDs, removeClaimIDs []uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, claimID := range removeClaimIDs {
		_, err := tx.ExecContext(ctx, `DELETE FROM tenant_role_claims WHERE tenant_role_id = $1 AND claim_id = $2`, tenantRoleID, claimID)
		if err != nil {
			return err
		}
	}

	for _, claimID := range assignClaimIDs {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO tenant_role_claims (id, tenant_role_id, claim_id, created_at)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (tenant_role_id, claim_id) DO NOTHING
		`, uuid.New(), tenantRoleID, claimID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
