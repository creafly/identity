package repository

import (
	"context"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	UpdateBlocked(ctx context.Context, id uuid.UUID, isBlocked bool, reason *string, blockedBy *uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.User, error)
	SetTOTPSecret(ctx context.Context, id uuid.UUID, secret string) error
	EnableTOTP(ctx context.Context, id uuid.UUID) error
	DisableTOTP(ctx context.Context, id uuid.UUID) error
	MarkEmailVerified(ctx context.Context, id uuid.UUID) error
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, first_name, last_name, locale, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Username, user.PasswordHash, user.FirstName, user.LastName,
		user.Locale, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, email, username, password_hash, first_name, last_name, avatar_url, locale, is_active, is_blocked, block_reason, blocked_at, blocked_by, email_verified, email_verified_at, totp_secret, totp_enabled, totp_verified_at, created_at, updated_at, deleted_at FROM users WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, email, username, password_hash, first_name, last_name, avatar_url, locale, is_active, is_blocked, block_reason, blocked_at, blocked_by, email_verified, email_verified_at, totp_secret, totp_enabled, totp_verified_at, created_at, updated_at, deleted_at FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, email, username, password_hash, first_name, last_name, avatar_url, locale, is_active, is_blocked, block_reason, blocked_at, blocked_by, email_verified, email_verified_at, totp_secret, totp_enabled, totp_verified_at, created_at, updated_at, deleted_at FROM users WHERE username = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users 
		SET email = $2, username = $3, first_name = $4, last_name = $5, avatar_url = $6, locale = $7, is_active = $8, updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Username, user.FirstName, user.LastName, user.AvatarURL, user.Locale, user.IsActive, user.UpdatedAt,
	)
	return err
}

func (r *userRepository) UpdateBlocked(ctx context.Context, id uuid.UUID, isBlocked bool, reason *string, blockedBy *uuid.UUID) error {
	var query string
	var err error
	if isBlocked {
		query = `UPDATE users SET is_blocked = true, block_reason = $2, blocked_by = $3, blocked_at = NOW(), updated_at = NOW() WHERE id = $1`
		_, err = r.db.ExecContext(ctx, query, id, reason, blockedBy)
	} else {
		query = `UPDATE users SET is_blocked = false, block_reason = NULL, blocked_by = NULL, blocked_at = NULL, updated_at = NOW() WHERE id = $1`
		_, err = r.db.ExecContext(ctx, query, id)
	}
	return err
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*entity.User, error) {
	var users []*entity.User
	query := `SELECT id, email, username, password_hash, first_name, last_name, avatar_url, locale, is_active, is_blocked, block_reason, blocked_at, blocked_by, email_verified, email_verified_at, totp_secret, totp_enabled, totp_verified_at, created_at, updated_at, deleted_at FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	return users, err
}

func (r *userRepository) SetTOTPSecret(ctx context.Context, id uuid.UUID, secret string) error {
	query := `UPDATE users SET totp_secret = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id, secret)
	return err
}

func (r *userRepository) EnableTOTP(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET totp_enabled = true, totp_verified_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *userRepository) DisableTOTP(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET totp_enabled = false, totp_secret = NULL, totp_verified_at = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *userRepository) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET email_verified = true, email_verified_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
