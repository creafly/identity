package repository

import (
	"context"
	"time"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	GetByIDIncludeDeleted(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int, includeDeleted bool) ([]*entity.Role, error)
	AssignToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error)
}

type roleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(db *sqlx.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	query := `
		INSERT INTO roles (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		role.ID, role.Name, role.Description, role.CreatedAt, role.UpdatedAt,
	)
	return err
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	var role entity.Role
	query := `SELECT id, name, description, created_at, updated_at, deleted_at FROM roles WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &role, query, id)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByIDIncludeDeleted(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	var role entity.Role
	query := `SELECT id, name, description, created_at, updated_at, deleted_at FROM roles WHERE id = $1`
	err := r.db.GetContext(ctx, &role, query, id)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role
	query := `SELECT id, name, description, created_at, updated_at, deleted_at FROM roles WHERE name = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &role, query, name)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	query := `
		UPDATE roles 
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, role.ID, role.Name, role.Description, role.UpdatedAt)
	return err
}

func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *roleRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET deleted_at = NULL, updated_at = $1 WHERE id = $2 AND deleted_at IS NOT NULL`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *roleRepository) List(ctx context.Context, offset, limit int, includeDeleted bool) ([]*entity.Role, error) {
	var roles []*entity.Role
	query := `SELECT id, name, description, created_at, updated_at, deleted_at FROM roles`
	if !includeDeleted {
		query += ` WHERE deleted_at IS NULL`
	}
	query += ` ORDER BY name OFFSET $1 LIMIT $2`
	err := r.db.SelectContext(ctx, &roles, query, offset, limit)
	return roles, err
}

func (r *roleRepository) AssignToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `
		INSERT INTO user_roles (id, user_id, role_id, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, role_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, utils.GenerateUUID(), userID, roleID)
	return err
}

func (r *roleRepository) RemoveFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	return err
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error) {
	var roles []*entity.Role
	query := `
		SELECT r.id, r.name, r.description, r.created_at, r.updated_at, r.deleted_at 
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.deleted_at IS NULL
	`
	err := r.db.SelectContext(ctx, &roles, query, userID)
	return roles, err
}
