package repository

import (
	"context"
	"time"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type EmailVerificationRepository interface {
	Create(ctx context.Context, token *entity.EmailVerificationToken) error
	GetLatestByUserID(ctx context.Context, userID uuid.UUID) (*entity.EmailVerificationToken, error)
	GetByCodeHash(ctx context.Context, userID uuid.UUID, codeHash string) (*entity.EmailVerificationToken, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
	DeleteExpiredTokens(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type emailVerificationRepository struct {
	db *sqlx.DB
}

func NewEmailVerificationRepository(db *sqlx.DB) EmailVerificationRepository {
	return &emailVerificationRepository{db: db}
}

func (r *emailVerificationRepository) Create(ctx context.Context, token *entity.EmailVerificationToken) error {
	query := `
		INSERT INTO email_verification_tokens (id, user_id, code_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.UserID, token.CodeHash, token.ExpiresAt, token.CreatedAt,
	)
	return err
}

func (r *emailVerificationRepository) GetLatestByUserID(ctx context.Context, userID uuid.UUID) (*entity.EmailVerificationToken, error) {
	var token entity.EmailVerificationToken
	query := `
		SELECT id, user_id, code_hash, expires_at, used_at, created_at 
		FROM email_verification_tokens 
		WHERE user_id = $1 AND used_at IS NULL AND expires_at > $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &token, query, userID, time.Now())
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *emailVerificationRepository) GetByCodeHash(ctx context.Context, userID uuid.UUID, codeHash string) (*entity.EmailVerificationToken, error) {
	var token entity.EmailVerificationToken
	query := `
		SELECT id, user_id, code_hash, expires_at, used_at, created_at 
		FROM email_verification_tokens 
		WHERE user_id = $1 AND code_hash = $2 AND used_at IS NULL AND expires_at > $3
	`
	err := r.db.GetContext(ctx, &token, query, userID, codeHash, time.Now())
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *emailVerificationRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE email_verification_tokens SET used_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *emailVerificationRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM email_verification_tokens WHERE expires_at < NOW() OR used_at IS NOT NULL`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *emailVerificationRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM email_verification_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
