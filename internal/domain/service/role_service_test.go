package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/creafly/identity/internal/testutil"
	"github.com/creafly/identity/internal/testutil/mocks"
)

func TestRoleService_Create(t *testing.T) {
	roleRepo := mocks.NewRoleRepositoryMock()
	svc := NewRoleService(roleRepo)
	ctx := context.Background()

	t.Run("valid role", func(t *testing.T) {
		input := CreateRoleInput{Name: "admin", Description: "Admin role"}

		role, err := svc.Create(ctx, input)
		if err != nil {
			t.Errorf("Create() error = %v", err)
			return
		}
		if role.Name != input.Name {
			t.Errorf("Create() name = %v, want %v", role.Name, input.Name)
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		existingRole := testutil.NewTestRole()
		roleRepo.AddRole(existingRole)

		input := CreateRoleInput{Name: existingRole.Name, Description: "Duplicate"}
		_, err := svc.Create(ctx, input)
		if err != ErrRoleAlreadyExists {
			t.Errorf("Create() error = %v, want %v", err, ErrRoleAlreadyExists)
		}
	})
}

func TestRoleService_GetByID(t *testing.T) {
	roleRepo := mocks.NewRoleRepositoryMock()
	svc := NewRoleService(roleRepo)
	ctx := context.Background()

	t.Run("existing role", func(t *testing.T) {
		role := testutil.NewTestRole()
		roleRepo.AddRole(role)

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
		if err != ErrRoleNotFound {
			t.Errorf("GetByID() error = %v, want %v", err, ErrRoleNotFound)
		}
	})
}

func TestRoleService_Update(t *testing.T) {
	roleRepo := mocks.NewRoleRepositoryMock()
	svc := NewRoleService(roleRepo)
	ctx := context.Background()

	t.Run("update role", func(t *testing.T) {
		role := testutil.NewTestRole()
		roleRepo.AddRole(role)

		newName := "updated-role"
		input := UpdateRoleInput{Name: &newName}
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
		input := UpdateRoleInput{Name: &newName}
		_, err := svc.Update(ctx, uuid.New(), input)
		if err != ErrRoleNotFound {
			t.Errorf("Update() error = %v, want %v", err, ErrRoleNotFound)
		}
	})
}

func TestRoleService_Delete(t *testing.T) {
	roleRepo := mocks.NewRoleRepositoryMock()
	svc := NewRoleService(roleRepo)
	ctx := context.Background()

	t.Run("delete existing role", func(t *testing.T) {
		role := testutil.NewTestRole()
		roleRepo.AddRole(role)

		err := svc.Delete(ctx, role.ID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}
	})

	t.Run("delete non-existing role", func(t *testing.T) {
		err := svc.Delete(ctx, uuid.New())
		if err != ErrRoleNotFound {
			t.Errorf("Delete() error = %v, want %v", err, ErrRoleNotFound)
		}
	})
}

func TestRoleService_AssignToUser(t *testing.T) {
	roleRepo := mocks.NewRoleRepositoryMock()
	svc := NewRoleService(roleRepo)
	ctx := context.Background()

	t.Run("assign existing role", func(t *testing.T) {
		role := testutil.NewTestRole()
		roleRepo.AddRole(role)
		userID := uuid.New()

		err := svc.AssignToUser(ctx, userID, role.ID)
		if err != nil {
			t.Errorf("AssignToUser() error = %v", err)
		}
	})

	t.Run("assign non-existing role", func(t *testing.T) {
		userID := uuid.New()
		err := svc.AssignToUser(ctx, userID, uuid.New())
		if err != ErrRoleNotFound {
			t.Errorf("AssignToUser() error = %v, want %v", err, ErrRoleNotFound)
		}
	})
}
