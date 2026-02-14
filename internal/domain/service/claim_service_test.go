package service

import (
	"context"
	"testing"

	"github.com/creafly/identity/internal/testutil"
	"github.com/creafly/identity/internal/testutil/mocks"
	"github.com/creafly/identity/internal/utils"
)

func TestClaimService_Create(t *testing.T) {
	claimRepo := mocks.NewClaimRepositoryMock()
	roleRepo := mocks.NewRoleRepositoryMock()
	tenantRoleRepo := mocks.NewTenantRoleRepositoryMock()
	svc := NewClaimService(claimRepo, roleRepo, tenantRoleRepo)
	ctx := context.Background()

	t.Run("valid claim", func(t *testing.T) {
		input := CreateClaimInput{Value: "test:permission:create"}
		claim, err := svc.Create(ctx, input)

		if err != nil {
			t.Errorf("Create() error = %v", err)
			return
		}
		if claim.Value != input.Value {
			t.Errorf("Create() value = %v, want %v", claim.Value, input.Value)
		}
	})

	t.Run("duplicate value", func(t *testing.T) {
		input := CreateClaimInput{Value: "test:permission:duplicate"}
		_, _ = svc.Create(ctx, input)

		_, err := svc.Create(ctx, input)
		if err != ErrClaimAlreadyExists {
			t.Errorf("Create() error = %v, want %v", err, ErrClaimAlreadyExists)
		}
	})
}

func TestClaimService_GetByID(t *testing.T) {
	claimRepo := mocks.NewClaimRepositoryMock()
	roleRepo := mocks.NewRoleRepositoryMock()
	tenantRoleRepo := mocks.NewTenantRoleRepositoryMock()
	svc := NewClaimService(claimRepo, roleRepo, tenantRoleRepo)
	ctx := context.Background()

	t.Run("existing claim", func(t *testing.T) {
		claim := testutil.NewTestClaim()
		claimRepo.AddClaim(claim)

		got, err := svc.GetByID(ctx, claim.ID)
		if err != nil {
			t.Errorf("GetByID() error = %v", err)
			return
		}
		if got.ID != claim.ID {
			t.Errorf("GetByID() ID = %v, want %v", got.ID, claim.ID)
		}
	})

	t.Run("non-existing claim", func(t *testing.T) {
		_, err := svc.GetByID(ctx, utils.GenerateUUID())
		if err != ErrClaimNotFound {
			t.Errorf("GetByID() error = %v, want %v", err, ErrClaimNotFound)
		}
	})
}

func TestClaimService_GetOrCreate(t *testing.T) {
	claimRepo := mocks.NewClaimRepositoryMock()
	roleRepo := mocks.NewRoleRepositoryMock()
	tenantRoleRepo := mocks.NewTenantRoleRepositoryMock()
	svc := NewClaimService(claimRepo, roleRepo, tenantRoleRepo)
	ctx := context.Background()

	t.Run("creates new claim", func(t *testing.T) {
		value := "test:new:claim"
		claim, err := svc.GetOrCreate(ctx, value)

		if err != nil {
			t.Errorf("GetOrCreate() error = %v", err)
			return
		}
		if claim.Value != value {
			t.Errorf("GetOrCreate() value = %v, want %v", claim.Value, value)
		}
	})

	t.Run("returns existing claim", func(t *testing.T) {
		existing := testutil.NewTestClaim()
		claimRepo.AddClaim(existing)

		claim, err := svc.GetOrCreate(ctx, existing.Value)
		if err != nil {
			t.Errorf("GetOrCreate() error = %v", err)
			return
		}
		if claim.ID != existing.ID {
			t.Errorf("GetOrCreate() ID = %v, want %v", claim.ID, existing.ID)
		}
	})
}

func TestClaimService_Delete(t *testing.T) {
	claimRepo := mocks.NewClaimRepositoryMock()
	roleRepo := mocks.NewRoleRepositoryMock()
	tenantRoleRepo := mocks.NewTenantRoleRepositoryMock()
	svc := NewClaimService(claimRepo, roleRepo, tenantRoleRepo)
	ctx := context.Background()

	t.Run("delete existing claim", func(t *testing.T) {
		claim := testutil.NewTestClaim()
		claimRepo.AddClaim(claim)

		err := svc.Delete(ctx, claim.ID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}
	})

	t.Run("delete non-existing claim", func(t *testing.T) {
		err := svc.Delete(ctx, utils.GenerateUUID())
		if err != ErrClaimNotFound {
			t.Errorf("Delete() error = %v, want %v", err, ErrClaimNotFound)
		}
	})
}

func TestClaimService_AssignToUser(t *testing.T) {
	claimRepo := mocks.NewClaimRepositoryMock()
	roleRepo := mocks.NewRoleRepositoryMock()
	tenantRoleRepo := mocks.NewTenantRoleRepositoryMock()
	svc := NewClaimService(claimRepo, roleRepo, tenantRoleRepo)
	ctx := context.Background()

	t.Run("assign existing claim to user", func(t *testing.T) {
		claim := testutil.NewTestClaim()
		claimRepo.AddClaim(claim)
		userID := utils.GenerateUUID()

		err := svc.AssignToUser(ctx, userID, claim.ID)
		if err != nil {
			t.Errorf("AssignToUser() error = %v", err)
		}
	})

	t.Run("assign non-existing claim", func(t *testing.T) {
		userID := utils.GenerateUUID()
		err := svc.AssignToUser(ctx, userID, utils.GenerateUUID())
		if err != ErrClaimNotFound {
			t.Errorf("AssignToUser() error = %v, want %v", err, ErrClaimNotFound)
		}
	})
}

func TestClaimService_AssignToRole(t *testing.T) {
	claimRepo := mocks.NewClaimRepositoryMock()
	roleRepo := mocks.NewRoleRepositoryMock()
	tenantRoleRepo := mocks.NewTenantRoleRepositoryMock()
	svc := NewClaimService(claimRepo, roleRepo, tenantRoleRepo)
	ctx := context.Background()

	t.Run("assign claim to role", func(t *testing.T) {
		claim := testutil.NewTestClaim()
		claimRepo.AddClaim(claim)

		role := testutil.NewTestRole()
		roleRepo.AddRole(role)

		err := svc.AssignToRole(ctx, role.ID, claim.ID)
		if err != nil {
			t.Errorf("AssignToRole() error = %v", err)
		}
	})

	t.Run("assign non-existing claim", func(t *testing.T) {
		role := testutil.NewTestRole()
		roleRepo.AddRole(role)

		err := svc.AssignToRole(ctx, role.ID, utils.GenerateUUID())
		if err != ErrClaimNotFound {
			t.Errorf("AssignToRole() error = %v, want %v", err, ErrClaimNotFound)
		}
	})

	t.Run("assign to non-existing role", func(t *testing.T) {
		claim := testutil.NewTestClaim()
		claimRepo.AddClaim(claim)

		err := svc.AssignToRole(ctx, utils.GenerateUUID(), claim.ID)
		if err != ErrRoleNotFound {
			t.Errorf("AssignToRole() error = %v, want %v", err, ErrRoleNotFound)
		}
	})
}

func TestClaimService_List(t *testing.T) {
	claimRepo := mocks.NewClaimRepositoryMock()
	roleRepo := mocks.NewRoleRepositoryMock()
	tenantRoleRepo := mocks.NewTenantRoleRepositoryMock()
	svc := NewClaimService(claimRepo, roleRepo, tenantRoleRepo)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		claimRepo.AddClaim(testutil.NewTestClaim())
	}

	t.Run("list with pagination", func(t *testing.T) {
		claims, err := svc.List(ctx, 0, 3)
		if err != nil {
			t.Errorf("List() error = %v", err)
			return
		}
		if len(claims) != 3 {
			t.Errorf("List() returned %d claims, want 3", len(claims))
		}
	})
}
