package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hexaend/identity/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

type PasswordResetRepository interface {
	Create(ctx context.Context, token *entity.PasswordResetToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.PasswordResetToken, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
	DeleteExpiredTokens(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type passwordResetRepository struct {
	db *sqlx.DB
}

func NewPasswordResetRepository(db *sqlx.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) Create(ctx context.Context, token *entity.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt,
	)
	return err
}

func (r *passwordResetRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.PasswordResetToken, error) {
	var token entity.PasswordResetToken
	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at 
		FROM password_reset_tokens 
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > $2
	`
	err := r.db.GetContext(ctx, &token, query, tokenHash, time.Now())
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *passwordResetRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *passwordResetRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM password_reset_tokens WHERE expires_at < NOW() OR used_at IS NOT NULL`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *passwordResetRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM password_reset_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
