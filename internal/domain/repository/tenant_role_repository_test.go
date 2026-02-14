package repository

import (
	"context"
	"testing"

	"github.com/creafly/identity/internal/testutil"
	"github.com/creafly/identity/internal/utils"
)

func TestTenantRoleRepository_Create(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewTenantRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("valid tenant role", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		role := testutil.NewTestTenantRole(tenant.ID)

		err := repo.Create(ctx, role)
		if err != nil {
			t.Errorf("Create() error = %v", err)
			return
		}

		got, err := repo.GetByID(ctx, role.ID)
		if err != nil {
			t.Errorf("GetByID() error = %v", err)
			return
		}
		if got.Name != role.Name {
			t.Errorf("Create() name = %v, want %v", got.Name, role.Name)
		}
	})
}

func TestTenantRoleRepository_GetByID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewTenantRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("existing role", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		role := testutil.NewTestTenantRole(tenant.ID)
		_ = repo.Create(ctx, role)

		got, err := repo.GetByID(ctx, role.ID)
		if err != nil {
			t.Errorf("GetByID() error = %v", err)
			return
		}
		if got.ID != role.ID {
			t.Errorf("GetByID() id = %v, want %v", got.ID, role.ID)
		}
	})

	t.Run("non-existing role", func(t *testing.T) {
		_, err := repo.GetByID(ctx, utils.GenerateUUID())
		if err == nil {
			t.Error("GetByID() expected error for non-existing role")
		}
	})
}

func TestTenantRoleRepository_GetByName(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewTenantRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("existing name", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		role := testutil.NewTestTenantRole(tenant.ID)
		_ = repo.Create(ctx, role)

		got, err := repo.GetByName(ctx, tenant.ID, role.Name)
		if err != nil {
			t.Errorf("GetByName() error = %v", err)
			return
		}
		if got.Name != role.Name {
			t.Errorf("GetByName() name = %v, want %v", got.Name, role.Name)
		}
	})

	t.Run("non-existing name", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		_, err := repo.GetByName(ctx, tenant.ID, "non-existing-role")
		if err == nil {
			t.Error("GetByName() expected error for non-existing name")
		}
	})
}

func TestTenantRoleRepository_Update(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewTenantRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("update role", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		role := testutil.NewTestTenantRole(tenant.ID)
		_ = repo.Create(ctx, role)

		role.Name = "updated-role-name"
		role.Description = "Updated description"
		err := repo.Update(ctx, role)
		if err != nil {
			t.Errorf("Update() error = %v", err)
			return
		}

		got, _ := repo.GetByID(ctx, role.ID)
		if got.Name != "updated-role-name" {
			t.Errorf("Update() name = %v, want updated-role-name", got.Name)
		}
	})
}

func TestTenantRoleRepository_Delete(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewTenantRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("delete non-default role", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		role := testutil.NewTestTenantRole(tenant.ID)
		role.IsDefault = false
		_ = repo.Create(ctx, role)

		err := repo.Delete(ctx, role.ID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
			return
		}

		_, err = repo.GetByID(ctx, role.ID)
		if err == nil {
			t.Error("Delete() role still exists after deletion")
		}
	})
}

func TestTenantRoleRepository_ListByTenant(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewTenantRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("list roles for tenant", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		role1 := testutil.NewTestTenantRole(tenant.ID)
		role2 := testutil.NewTestTenantRole(tenant.ID)
		_ = repo.Create(ctx, role1)
		_ = repo.Create(ctx, role2)

		roles, err := repo.ListByTenant(ctx, tenant.ID, false)
		if err != nil {
			t.Errorf("ListByTenant() error = %v", err)
			return
		}
		if len(roles) < 2 {
			t.Errorf("ListByTenant() count = %d, want >= 2", len(roles))
		}
	})
}

func TestTenantRoleRepository_AddClaim(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewTenantRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("add claim to role", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		role := testutil.NewTestTenantRole(tenant.ID)
		_ = repo.Create(ctx, role)
		claim := tdb.CreateTestClaim(ctx, t)

		err := repo.AddClaim(ctx, role.ID, claim.ID)
		if err != nil {
			t.Errorf("AddClaim() error = %v", err)
			return
		}

		claims, _ := repo.GetRoleClaims(ctx, role.ID)
		if len(claims) != 1 {
			t.Errorf("AddClaim() claims count = %d, want 1", len(claims))
		}
	})

	t.Run("add same claim twice (idempotent)", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		role := testutil.NewTestTenantRole(tenant.ID)
		_ = repo.Create(ctx, role)
		claim := tdb.CreateTestClaim(ctx, t)

		_ = repo.AddClaim(ctx, role.ID, claim.ID)
		err := repo.AddClaim(ctx, role.ID, claim.ID)
		if err != nil {
			t.Errorf("AddClaim() second call error = %v", err)
		}
	})
}

func TestTenantRoleRepository_RemoveClaim(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewTenantRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("remove claim from role", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		role := testutil.NewTestTenantRole(tenant.ID)
		_ = repo.Create(ctx, role)
		claim := tdb.CreateTestClaim(ctx, t)
		_ = repo.AddClaim(ctx, role.ID, claim.ID)

		err := repo.RemoveClaim(ctx, role.ID, claim.ID)
		if err != nil {
			t.Errorf("RemoveClaim() error = %v", err)
			return
		}

		claims, _ := repo.GetRoleClaims(ctx, role.ID)
		if len(claims) != 0 {
			t.Errorf("RemoveClaim() claims count = %d, want 0", len(claims))
		}
	})
}

func TestTenantRoleRepository_AssignToUser(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewTenantRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("assign role to user", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		user := tdb.CreateTestUser(ctx, t)
		role := testutil.NewTestTenantRole(tenant.ID)
		_ = repo.Create(ctx, role)

		err := repo.AssignToUser(ctx, user.ID, tenant.ID, role.ID)
		if err != nil {
			t.Errorf("AssignToUser() error = %v", err)
			return
		}

		roles, _ := repo.GetUserRoles(ctx, user.ID, tenant.ID)
		if len(roles) != 1 {
			t.Errorf("AssignToUser() roles count = %d, want 1", len(roles))
		}
	})

	t.Run("assign same role twice (idempotent)", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		user := tdb.CreateTestUser(ctx, t)
		role := testutil.NewTestTenantRole(tenant.ID)
		_ = repo.Create(ctx, role)

		_ = repo.AssignToUser(ctx, user.ID, tenant.ID, role.ID)
		err := repo.AssignToUser(ctx, user.ID, tenant.ID, role.ID)
		if err != nil {
			t.Errorf("AssignToUser() second call error = %v", err)
		}
	})
}

func TestTenantRoleRepository_RemoveFromUser(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewTenantRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("remove role from user", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		user := tdb.CreateTestUser(ctx, t)
		role := testutil.NewTestTenantRole(tenant.ID)
		_ = repo.Create(ctx, role)
		_ = repo.AssignToUser(ctx, user.ID, tenant.ID, role.ID)

		err := repo.RemoveFromUser(ctx, user.ID, tenant.ID, role.ID)
		if err != nil {
			t.Errorf("RemoveFromUser() error = %v", err)
			return
		}

		roles, _ := repo.GetUserRoles(ctx, user.ID, tenant.ID)
		if len(roles) != 0 {
			t.Errorf("RemoveFromUser() roles count = %d, want 0", len(roles))
		}
	})
}

func TestTenantRoleRepository_GetUserClaims(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewTenantRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("user with role claims", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		user := tdb.CreateTestUser(ctx, t)
		role := testutil.NewTestTenantRole(tenant.ID)
		_ = repo.Create(ctx, role)
		claim1 := tdb.CreateTestClaim(ctx, t)
		claim2 := tdb.CreateTestClaim(ctx, t)
		_ = repo.AddClaim(ctx, role.ID, claim1.ID)
		_ = repo.AddClaim(ctx, role.ID, claim2.ID)
		_ = repo.AssignToUser(ctx, user.ID, tenant.ID, role.ID)

		claims, err := repo.GetUserClaims(ctx, user.ID, tenant.ID)
		if err != nil {
			t.Errorf("GetUserClaims() error = %v", err)
			return
		}
		if len(claims) != 2 {
			t.Errorf("GetUserClaims() count = %d, want 2", len(claims))
		}
	})

	t.Run("user with no roles", func(t *testing.T) {
		tenant := tdb.CreateTestTenant(ctx, t)
		user := tdb.CreateTestUser(ctx, t)

		claims, err := repo.GetUserClaims(ctx, user.ID, tenant.ID)
		if err != nil {
			t.Errorf("GetUserClaims() error = %v", err)
			return
		}
		if len(claims) != 0 {
			t.Errorf("GetUserClaims() count = %d, want 0", len(claims))
		}
	})
}
