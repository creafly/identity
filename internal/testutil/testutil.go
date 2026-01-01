package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/creafly/identity/internal/domain/entity"
	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type TestDB struct {
	DB *sqlx.DB
}

func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5440/identity_test?sslmode=disable"
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		t.Skipf("Skipping integration test: could not connect to test database: %v", err)
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	return &TestDB{DB: db}
}

func (tdb *TestDB) Cleanup(t *testing.T) {
	t.Helper()

	if tdb.DB != nil {
		tables := []string{
			"user_tenant_roles",
			"tenant_role_claims",
			"tenant_roles",
			"tenant_members",
			"role_claims",
			"user_claims",
			"user_roles",
			"password_reset_tokens",
			"outbox_events",
		}

		for _, table := range tables {
			_, _ = tdb.DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		}

		_, _ = tdb.DB.Exec("DELETE FROM claims WHERE id NOT IN (SELECT id FROM claims WHERE value LIKE '%:%')")
		_, _ = tdb.DB.Exec("DELETE FROM roles WHERE id NOT IN ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002')")
		_, _ = tdb.DB.Exec("DELETE FROM users")
		_, _ = tdb.DB.Exec("DELETE FROM tenants")

		tdb.DB.Close()
	}
}

func (tdb *TestDB) CleanupTables(t *testing.T, tables ...string) {
	t.Helper()

	for _, table := range tables {
		_, err := tdb.DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("warning: failed to clean table %s: %v", table, err)
		}
	}
}

func runMigrations(db *sqlx.DB) error {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get current file path")
	}

	testutilDir := filepath.Dir(currentFile)
	internalDir := filepath.Dir(testutilDir)
	identityDir := filepath.Dir(internalDir)
	migrationsPath := filepath.Join(identityDir, "migrations")

	migrationsPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func NewTestUser() *entity.User {
	return &entity.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		PasswordHash: "$2a$10$test.hash.value",
		FirstName:    "Test",
		LastName:     "User",
		Locale:       "en-US",
		IsActive:     true,
		IsBlocked:    false,
		TotpEnabled:  false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func NewTestTenant() *entity.Tenant {
	id := uuid.New()
	return &entity.Tenant{
		ID:          id,
		Name:        fmt.Sprintf("Test Tenant %s", id.String()[:8]),
		DisplayName: fmt.Sprintf("Test Display Name %s", id.String()[:8]),
		Slug:        fmt.Sprintf("test-tenant-%s", id.String()[:8]),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func NewTestRole() *entity.Role {
	id := uuid.New()
	return &entity.Role{
		ID:          id,
		Name:        fmt.Sprintf("test-role-%s", id.String()[:8]),
		Description: "Test role description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func NewTestClaim() *entity.Claim {
	id := uuid.New()
	return &entity.Claim{
		ID:        id,
		Value:     fmt.Sprintf("test:claim:%s", id.String()[:8]),
		CreatedAt: time.Now(),
	}
}

func NewTestTenantRole(tenantID uuid.UUID) *entity.TenantRole {
	id := uuid.New()
	return &entity.TenantRole{
		ID:          id,
		TenantID:    tenantID,
		Name:        fmt.Sprintf("test-tenant-role-%s", id.String()[:8]),
		Description: "Test tenant role description",
		IsDefault:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func NewTestOutboxEvent() *entity.OutboxEvent {
	return &entity.OutboxEvent{
		ID:         uuid.New(),
		EventType:  "test.event",
		Payload:    `{"test": "data"}`,
		Status:     "pending",
		RetryCount: 0,
		CreatedAt:  time.Now(),
	}
}

func NewTestPasswordResetToken(userID uuid.UUID) *entity.PasswordResetToken {
	return &entity.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: fmt.Sprintf("hash-%s", uuid.New().String()),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}
}

func (tdb *TestDB) CreateTestUser(ctx context.Context, t *testing.T) *entity.User {
	t.Helper()

	user := NewTestUser()
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, locale, is_active, is_blocked, totp_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := tdb.DB.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.Locale, user.IsActive, user.IsBlocked, user.TotpEnabled,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return user
}

func (tdb *TestDB) CreateTestTenant(ctx context.Context, t *testing.T) *entity.Tenant {
	t.Helper()

	tenant := NewTestTenant()
	query := `
		INSERT INTO tenants (id, name, display_name, slug, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tdb.DB.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.DisplayName, tenant.Slug,
		tenant.IsActive, tenant.CreatedAt, tenant.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("failed to create test tenant: %v", err)
	}

	return tenant
}

func (tdb *TestDB) CreateTestClaim(ctx context.Context, t *testing.T) *entity.Claim {
	t.Helper()

	claim := NewTestClaim()
	query := `INSERT INTO claims (id, value, created_at) VALUES ($1, $2, $3)`
	_, err := tdb.DB.ExecContext(ctx, query, claim.ID, claim.Value, claim.CreatedAt)
	if err != nil {
		t.Fatalf("failed to create test claim: %v", err)
	}

	return claim
}

func (tdb *TestDB) CreateTestRole(ctx context.Context, t *testing.T) *entity.Role {
	t.Helper()

	role := NewTestRole()
	query := `INSERT INTO roles (id, name, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := tdb.DB.ExecContext(ctx, query, role.ID, role.Name, role.Description, role.CreatedAt, role.UpdatedAt)
	if err != nil {
		t.Fatalf("failed to create test role: %v", err)
	}

	return role
}
