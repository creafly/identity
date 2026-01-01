package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/creafly/identity/internal/testutil"
	"github.com/creafly/identity/internal/testutil/mocks"
)

func TestTenantRoleService_Create(t *testing.T) {
	tenantRoleRepo := mocks.NewTenantRoleRepositoryMock()
	claimRepo := mocks.NewClaimRepositoryMock()
	svc := NewTenantRoleService(tenantRoleRepo, claimRepo)
	ctx := context.Background()

	t.Run("valid tenant role", func(t *testing.T) {
		tenantID := uuid.New()
		input := CreateTenantRoleInput{
			TenantID:    tenantID,
			Name:        "editor",
			Description: "Editor role",
		}

		role, err := svc.Create(ctx, input)
		if err != nil {
			t.Errorf("Create() error = %v", err)
			return
		}
		if role.Name != input.Name {
			t.Errorf("Create() name = %v, want %v", role.Name, input.Name)
		}
		if role.TenantID != tenantID {
			t.Errorf("Create() tenantID = %v, want %v", role.TenantID, tenantID)
		}
	})

	t.Run("duplicate name in same tenant", func(t *testing.T) {
		tenantID := uuid.New()
		existingRole := testutil.NewTestTenantRole(tenantID)
		tenantRoleRepo.AddTenantRole(existingRole)

		input := CreateTenantRoleInput{
			TenantID:    tenantID,
			Name:        existingRole.Name,
			Description: "Duplicate",
		}
		_, err := svc.Create(ctx, input)
		if err != ErrTenantRoleAlreadyExists {
			t.Errorf("Create() error = %v, want %v", err, ErrTenantRoleAlreadyExists)
		}
	})
}

func TestTenantRoleService_GetByID(t *testing.T) {
	tenantRoleRepo := mocks.NewTenantRoleRepositoryMock()
	claimRepo := mocks.NewClaimRepositoryMock()
	svc := NewTenantRoleService(tenantRoleRepo, claimRepo)
	ctx := context.Background()

	t.Run("existing role", func(t *testing.T) {
		role := testutil.NewTestTenantRole(uuid.New())
		tenantRoleRepo.AddTenantRole(role)

		got, err := svc.GetByID(ctx, role.ID)
		if err != nil {
			t.Errorf("GetByID() error = %v", err)
			return
		}
		if got.ID != role.ID {
			t.Errorf("GetByID() id = %v, want %v", got.ID, role.ID)
		}
	})

	t.Run("non-existing role", func(t *testing.T) {
		_, err := svc.GetByID(ctx, uuid.New())
		if err != ErrTenantRoleNotFound {
			t.Errorf("GetByID() error = %v, want %v", err, ErrTenantRoleNotFound)
		}
	})
}

func TestTenantRoleService_Update(t *testing.T) {
	tenantRoleRepo := mocks.NewTenantRoleRepositoryMock()
	claimRepo := mocks.NewClaimRepositoryMock()
	svc := NewTenantRoleService(tenantRoleRepo, claimRepo)
	ctx := context.Background()

	t.Run("update role", func(t *testing.T) {
		role := testutil.NewTestTenantRole(uuid.New())
		tenantRoleRepo.AddTenantRole(role)

		newName := "updated-tenant-role"
		input := UpdateTenantRoleInput{Name: &newName}
		updated, err := svc.Update(ctx, role.ID, input)
		if err != nil {
			t.Errorf("Update() error = %v", err)
			return
		}
		if updated.Name != newName {
			t.Errorf("Update() name = %v, want %v", updated.Name, newName)
		}
	})

	t.Run("update non-existing role", func(t *testing.T) {
		newName := "test"
		input := UpdateTenantRoleInput{Name: &newName}
		_, err := svc.Update(ctx, uuid.New(), input)
		if err != ErrTenantRoleNotFound {
			t.Errorf("Update() error = %v, want %v", err, ErrTenantRoleNotFound)
		}
	})
}

func TestTenantRoleService_Delete(t *testing.T) {
	tenantRoleRepo := mocks.NewTenantRoleRepositoryMock()
	claimRepo := mocks.NewClaimRepositoryMock()
	svc := NewTenantRoleService(tenantRoleRepo, claimRepo)
	ctx := context.Background()

	t.Run("delete non-default role", func(t *testing.T) {
		role := testutil.NewTestTenantRole(uuid.New())
		role.IsDefault = false
		tenantRoleRepo.AddTenantRole(role)

		err := svc.Delete(ctx, role.ID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}
	})

	t.Run("delete non-existing role", func(t *testing.T) {
		err := svc.Delete(ctx, uuid.New())
		if err != ErrTenantRoleNotFound {
			t.Errorf("Delete() error = %v, want %v", err, ErrTenantRoleNotFound)
		}
	})

	t.Run("delete default role should fail", func(t *testing.T) {
		role := testutil.NewTestTenantRole(uuid.New())
		role.IsDefault = true
		tenantRoleRepo.AddTenantRole(role)

		err := svc.Delete(ctx, role.ID)
		if err != ErrCannotDeleteDefaultRole {
			t.Errorf("Delete() error = %v, want %v", err, ErrCannotDeleteDefaultRole)
		}
	})
}

func TestTenantRoleService_AssignToUser(t *testing.T) {
	tenantRoleRepo := mocks.NewTenantRoleRepositoryMock()
	claimRepo := mocks.NewClaimRepositoryMock()
	svc := NewTenantRoleService(tenantRoleRepo, claimRepo)
	ctx := context.Background()

	t.Run("assign existing role", func(t *testing.T) {
		tenantID := uuid.New()
		role := testutil.NewTestTenantRole(tenantID)
		tenantRoleRepo.AddTenantRole(role)
		userID := uuid.New()

		err := svc.AssignToUser(ctx, userID, tenantID, role.ID)
		if err != nil {
			t.Errorf("AssignToUser() error = %v", err)
		}
	})

	t.Run("assign non-existing role", func(t *testing.T) {
		err := svc.AssignToUser(ctx, uuid.New(), uuid.New(), uuid.New())
		if err != ErrTenantRoleNotFound {
			t.Errorf("AssignToUser() error = %v, want %v", err, ErrTenantRoleNotFound)
		}
	})
}
