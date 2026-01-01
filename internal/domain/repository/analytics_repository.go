package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type UserAnalytics struct {
	TotalUsers        int `db:"total_users" json:"totalUsers"`
	ActiveUsers       int `db:"active_users" json:"activeUsers"`
	BlockedUsers      int `db:"blocked_users" json:"blockedUsers"`
	NewUsersThisMonth int `db:"new_users_this_month" json:"newUsersThisMonth"`
	MAU               int `db:"mau" json:"mau"`
	DAU               int `db:"dau" json:"dau"`
	TotalTenants      int `db:"total_tenants" json:"totalTenants"`
}

type AnalyticsRepository interface {
	GetUserAnalytics(ctx context.Context) (*UserAnalytics, error)
}

type analyticsRepository struct {
	db *sqlx.DB
}

func NewAnalyticsRepository(db *sqlx.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) GetUserAnalytics(ctx context.Context) (*UserAnalytics, error) {
	var analytics UserAnalytics

	query := `
		SELECT
			COALESCE((SELECT COUNT(*) FROM users WHERE deleted_at IS NULL), 0) as total_users,
			COALESCE((SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND is_active = true AND is_blocked = false), 0) as active_users,
			COALESCE((SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND is_blocked = true), 0) as blocked_users,
			COALESCE((SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND created_at >= date_trunc('month', CURRENT_DATE)), 0) as new_users_this_month,
			COALESCE((SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND updated_at >= date_trunc('month', CURRENT_DATE)), 0) as mau,
			COALESCE((SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND updated_at >= CURRENT_DATE), 0) as dau,
			COALESCE((SELECT COUNT(*) FROM tenants WHERE deleted_at IS NULL), 0) as total_tenants
	`

	err := r.db.GetContext(ctx, &analytics, query)
	if err != nil {
		return nil, err
	}

	return &analytics, nil
}
