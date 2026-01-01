package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/creafly/identity/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

type ClaimRepository interface {
	Create(ctx context.Context, claim *entity.Claim) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Claim, error)
	GetByValue(ctx context.Context, value string) (*entity.Claim, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.Claim, error)
	AssignToUser(ctx context.Context, userID, claimID uuid.UUID) error
	RemoveFromUser(ctx context.Context, userID, claimID uuid.UUID) error
	GetUserClaims(ctx context.Context, userID uuid.UUID) ([]*entity.Claim, error)
	AssignToRole(ctx context.Context, roleID, claimID uuid.UUID) error
	RemoveFromRole(ctx context.Context, roleID, claimID uuid.UUID) error
	GetRoleClaims(ctx context.Context, roleID uuid.UUID) ([]*entity.Claim, error)
}

type claimRepository struct {
	db *sqlx.DB
}

func NewClaimRepository(db *sqlx.DB) ClaimRepository {
	return &claimRepository{db: db}
}

func (r *claimRepository) Create(ctx context.Context, claim *entity.Claim) error {
	query := `
		INSERT INTO claims (id, value, created_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query,
		claim.ID, claim.Value, claim.CreatedAt,
	)
	return err
}

func (r *claimRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Claim, error) {
	var claim entity.Claim
	query := `SELECT id, value, created_at FROM claims WHERE id = $1`
	err := r.db.GetContext(ctx, &claim, query, id)
	if err != nil {
		return nil, err
	}
	return &claim, nil
}

func (r *claimRepository) GetByValue(ctx context.Context, value string) (*entity.Claim, error) {
	var claim entity.Claim
	query := `SELECT id, value, created_at FROM claims WHERE value = $1`
	err := r.db.GetContext(ctx, &claim, query, value)
	if err != nil {
		return nil, err
	}
	return &claim, nil
}

func (r *claimRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM claims WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *claimRepository) List(ctx context.Context, offset, limit int) ([]*entity.Claim, error) {
	var claims []*entity.Claim
	query := `SELECT id, value, created_at FROM claims ORDER BY value LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &claims, query, limit, offset)
	return claims, err
}

func (r *claimRepository) AssignToUser(ctx context.Context, userID, claimID uuid.UUID) error {
	query := `
		INSERT INTO user_claims (id, user_id, claim_id, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, claim_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, uuid.New(), userID, claimID)
	return err
}

func (r *claimRepository) RemoveFromUser(ctx context.Context, userID, claimID uuid.UUID) error {
	query := `DELETE FROM user_claims WHERE user_id = $1 AND claim_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, claimID)
	return err
}

func (r *claimRepository) GetUserClaims(ctx context.Context, userID uuid.UUID) ([]*entity.Claim, error) {
	var claims []*entity.Claim
	query := `
		SELECT c.id, c.value, c.created_at FROM claims c
		JOIN user_claims uc ON uc.claim_id = c.id
		WHERE uc.user_id = $1
		ORDER BY c.value
	`
	err := r.db.SelectContext(ctx, &claims, query, userID)
	return claims, err
}

func (r *claimRepository) AssignToRole(ctx context.Context, roleID, claimID uuid.UUID) error {
	query := `
		INSERT INTO role_claims (id, role_id, claim_id, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (role_id, claim_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, uuid.New(), roleID, claimID)
	return err
}

func (r *claimRepository) RemoveFromRole(ctx context.Context, roleID, claimID uuid.UUID) error {
	query := `DELETE FROM role_claims WHERE role_id = $1 AND claim_id = $2`
	_, err := r.db.ExecContext(ctx, query, roleID, claimID)
	return err
}

func (r *claimRepository) GetRoleClaims(ctx context.Context, roleID uuid.UUID) ([]*entity.Claim, error) {
	var claims []*entity.Claim
	query := `
		SELECT c.id, c.value, c.created_at FROM claims c
		JOIN role_claims rc ON rc.claim_id = c.id
		WHERE rc.role_id = $1
		ORDER BY c.value
	`
	err := r.db.SelectContext(ctx, &claims, query, roleID)
	return claims, err
}
