package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	Email           string     `db:"email" json:"email"`
	Username        *string    `db:"username" json:"username,omitempty"`
	PasswordHash    string     `db:"password_hash" json:"-"`
	FirstName       string     `db:"first_name" json:"firstName"`
	LastName        string     `db:"last_name" json:"lastName"`
	AvatarURL       *string    `db:"avatar_url" json:"avatarUrl,omitempty"`
	Locale          string     `db:"locale" json:"locale"`
	IsActive        bool       `db:"is_active" json:"isActive"`
	IsBlocked       bool       `db:"is_blocked" json:"isBlocked"`
	BlockReason     *string    `db:"block_reason" json:"blockReason,omitempty"`
	BlockedAt       *time.Time `db:"blocked_at" json:"blockedAt,omitempty"`
	BlockedBy       *uuid.UUID `db:"blocked_by" json:"blockedBy,omitempty"`
	EmailVerified   bool       `db:"email_verified" json:"emailVerified"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"emailVerifiedAt,omitempty"`
	TotpSecret      *string    `db:"totp_secret" json:"-"`
	TotpEnabled     bool       `db:"totp_enabled" json:"totpEnabled"`
	TotpVerified    *time.Time `db:"totp_verified_at" json:"totpVerifiedAt,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt       *time.Time `db:"deleted_at" json:"deletedAt,omitempty"`
}

type Tenant struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	DisplayName string     `db:"display_name" json:"displayName"`
	Slug        string     `db:"slug" json:"slug"`
	IsActive    bool       `db:"is_active" json:"isActive"`
	IsBlocked   bool       `db:"is_blocked" json:"isBlocked"`
	BlockReason *string    `db:"block_reason" json:"blockReason,omitempty"`
	BlockedAt   *time.Time `db:"blocked_at" json:"blockedAt,omitempty"`
	BlockedBy   *uuid.UUID `db:"blocked_by" json:"blockedBy,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deletedAt,omitempty"`
}

type Role struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}

type Claim struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Value     string    `db:"value" json:"value"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type UserRole struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"userId"`
	RoleID    uuid.UUID `db:"role_id" json:"roleId"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type UserClaim struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"userId"`
	ClaimID   uuid.UUID `db:"claim_id" json:"claimId"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type RoleClaim struct {
	ID        uuid.UUID `db:"id" json:"id"`
	RoleID    uuid.UUID `db:"role_id" json:"roleId"`
	ClaimID   uuid.UUID `db:"claim_id" json:"claimId"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type TenantMember struct {
	ID        uuid.UUID `db:"id" json:"id"`
	TenantID  uuid.UUID `db:"tenant_id" json:"tenantId"`
	UserID    uuid.UUID `db:"user_id" json:"userId"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type TenantRole struct {
	ID          uuid.UUID `db:"id" json:"id"`
	TenantID    uuid.UUID `db:"tenant_id" json:"tenantId"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	IsDefault   bool      `db:"is_default" json:"isDefault"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}

type TenantRoleClaim struct {
	ID           uuid.UUID `db:"id" json:"id"`
	TenantRoleID uuid.UUID `db:"tenant_role_id" json:"tenantRoleId"`
	ClaimID      uuid.UUID `db:"claim_id" json:"claimId"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}

type UserTenantRole struct {
	ID           uuid.UUID `db:"id" json:"id"`
	UserID       uuid.UUID `db:"user_id" json:"userId"`
	TenantID     uuid.UUID `db:"tenant_id" json:"tenantId"`
	TenantRoleID uuid.UUID `db:"tenant_role_id" json:"tenantRoleId"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}

const (
	SystemRoleAdminID = "00000000-0000-0000-0000-000000000001"
	SystemRoleUserID  = "00000000-0000-0000-0000-000000000002"
)

type PasswordResetToken struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    uuid.UUID  `db:"user_id" json:"userId"`
	TokenHash string     `db:"token_hash" json:"-"`
	ExpiresAt time.Time  `db:"expires_at" json:"expiresAt"`
	UsedAt    *time.Time `db:"used_at" json:"usedAt,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
}

type EmailVerificationToken struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    uuid.UUID  `db:"user_id" json:"userId"`
	CodeHash  string     `db:"code_hash" json:"-"`
	ExpiresAt time.Time  `db:"expires_at" json:"expiresAt"`
	UsedAt    *time.Time `db:"used_at" json:"usedAt,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
}
