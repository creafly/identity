package repository

import (
	"context"
	"testing"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/testutil"
)

func TestRoleRepository_Create(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("valid role", func(t *testing.T) {
		role := testutil.NewTestRole()
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

func TestRoleRepository_GetByID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("existing role", func(t *testing.T) {
		role := tdb.CreateTestRole(ctx, t)

		got, err := repo.GetByID(ctx, role.ID)
		if err != nil {
			t.Errorf("GetByID() error = %v", err)
			return
		}
		if got.ID != role.ID {
			t.Errorf("GetByID() ID = %v, want %v", got.ID, role.ID)
		}
	})

	t.Run("non-existing role", func(t *testing.T) {
		_, err := repo.GetByID(ctx, testutil.NewTestRole().ID)
		if err == nil {
			t.Error("GetByID() expected error for non-existing role")
		}
	})
}

func TestRoleRepository_GetByName(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("existing name", func(t *testing.T) {
		role := tdb.CreateTestRole(ctx, t)

		got, err := repo.GetByName(ctx, role.Name)
		if err != nil {
			t.Errorf("GetByName() error = %v", err)
			return
		}
		if got.Name != role.Name {
			t.Errorf("GetByName() = %v, want %v", got.Name, role.Name)
		}
	})

	t.Run("non-existing name", func(t *testing.T) {
		_, err := repo.GetByName(ctx, "non-existing-role")
		if err == nil {
			t.Error("GetByName() expected error for non-existing name")
		}
	})
}

func TestRoleRepository_Update(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("update role", func(t *testing.T) {
		role := tdb.CreateTestRole(ctx, t)
		role.Description = "Updated description"

		err := repo.Update(ctx, role)
		if err != nil {
			t.Errorf("Update() error = %v", err)
			return
		}

		got, _ := repo.GetByID(ctx, role.ID)
		if got.Description != "Updated description" {
			t.Errorf("Update() description = %v, want %v", got.Description, "Updated description")
		}
	})
}

func TestRoleRepository_Delete(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("delete existing role", func(t *testing.T) {
		role := tdb.CreateTestRole(ctx, t)

		err := repo.Delete(ctx, role.ID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
			return
		}

		_, err = repo.GetByID(ctx, role.ID)
		if err == nil {
			t.Error("GetByID() expected error after delete")
		}
	})
}

func TestRoleRepository_AssignToUser(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("assign role to user", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		role := tdb.CreateTestRole(ctx, t)

		err := repo.AssignToUser(ctx, user.ID, role.ID)
		if err != nil {
			t.Errorf("AssignToUser() error = %v", err)
			return
		}

		roles, err := repo.GetUserRoles(ctx, user.ID)
		if err != nil {
			t.Errorf("GetUserRoles() error = %v", err)
			return
		}
		if len(roles) != 1 {
			t.Errorf("GetUserRoles() returned %d roles, want 1", len(roles))
		}
	})
}

func TestRoleRepository_RemoveFromUser(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("remove role from user", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		role := tdb.CreateTestRole(ctx, t)

		_ = repo.AssignToUser(ctx, user.ID, role.ID)

		err := repo.RemoveFromUser(ctx, user.ID, role.ID)
		if err != nil {
			t.Errorf("RemoveFromUser() error = %v", err)
			return
		}

		roles, _ := repo.GetUserRoles(ctx, user.ID)
		if len(roles) != 0 {
			t.Errorf("GetUserRoles() returned %d roles after remove, want 0", len(roles))
		}
	})
}

func TestRoleRepository_GetUserRoles(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewRoleRepository(tdb.DB)
	ctx := context.Background()

	t.Run("user with multiple roles", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		role1 := tdb.CreateTestRole(ctx, t)
		role2 := tdb.CreateTestRole(ctx, t)

		_ = repo.AssignToUser(ctx, user.ID, role1.ID)
		_ = repo.AssignToUser(ctx, user.ID, role2.ID)

		roles, err := repo.GetUserRoles(ctx, user.ID)
		if err != nil {
			t.Errorf("GetUserRoles() error = %v", err)
			return
		}
		if len(roles) != 2 {
			t.Errorf("GetUserRoles() returned %d roles, want 2", len(roles))
		}
	})

	t.Run("user with no roles", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)

		roles, err := repo.GetUserRoles(ctx, user.ID)
		if err != nil {
			t.Errorf("GetUserRoles() error = %v", err)
			return
		}
		if len(roles) != 0 {
			t.Errorf("GetUserRoles() returned %d roles, want 0", len(roles))
		}
	})
}

var _ entity.Role
