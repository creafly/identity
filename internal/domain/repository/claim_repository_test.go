package repository

import (
	"context"
	"testing"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/testutil"
)

func TestClaimRepository_Create(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewClaimRepository(tdb.DB)
	ctx := context.Background()

	t.Run("valid claim", func(t *testing.T) {
		claim := testutil.NewTestClaim()
		err := repo.Create(ctx, claim)

		if err != nil {
			t.Errorf("Create() error = %v", err)
			return
		}

		got, err := repo.GetByID(ctx, claim.ID)
		if err != nil {
			t.Errorf("GetByID() error = %v", err)
			return
		}
		if got.Value != claim.Value {
			t.Errorf("Create() value = %v, want %v", got.Value, claim.Value)
		}
	})

	t.Run("duplicate value should fail", func(t *testing.T) {
		claim1 := testutil.NewTestClaim()

		err := repo.Create(ctx, claim1)
		if err != nil {
			t.Fatalf("First Create() error = %v", err)
		}

		claim2 := testutil.NewTestClaim()
		claim2.Value = claim1.Value

		err = repo.Create(ctx, claim2)
		if err == nil {
			t.Error("Create() expected error for duplicate value, got nil")
		}
	})
}

func TestClaimRepository_GetByID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewClaimRepository(tdb.DB)
	ctx := context.Background()

	t.Run("existing claim", func(t *testing.T) {
		claim := tdb.CreateTestClaim(ctx, t)

		got, err := repo.GetByID(ctx, claim.ID)
		if err != nil {
			t.Errorf("GetByID() error = %v", err)
			return
		}
		if got.ID != claim.ID {
			t.Errorf("GetByID() ID = %v, want %v", got.ID, claim.ID)
		}
	})

	t.Run("non-existing claim", func(t *testing.T) {
		_, err := repo.GetByID(ctx, testutil.NewTestClaim().ID)
		if err == nil {
			t.Error("GetByID() expected error for non-existing claim")
		}
	})
}

func TestClaimRepository_GetByValue(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewClaimRepository(tdb.DB)
	ctx := context.Background()

	t.Run("existing value", func(t *testing.T) {
		claim := tdb.CreateTestClaim(ctx, t)

		got, err := repo.GetByValue(ctx, claim.Value)
		if err != nil {
			t.Errorf("GetByValue() error = %v", err)
			return
		}
		if got.Value != claim.Value {
			t.Errorf("GetByValue() = %v, want %v", got.Value, claim.Value)
		}
	})

	t.Run("non-existing value", func(t *testing.T) {
		_, err := repo.GetByValue(ctx, "non:existing:value")
		if err == nil {
			t.Error("GetByValue() expected error for non-existing value")
		}
	})
}

func TestClaimRepository_Delete(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewClaimRepository(tdb.DB)
	ctx := context.Background()

	t.Run("delete existing claim", func(t *testing.T) {
		claim := tdb.CreateTestClaim(ctx, t)

		err := repo.Delete(ctx, claim.ID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
			return
		}

		_, err = repo.GetByID(ctx, claim.ID)
		if err == nil {
			t.Error("GetByID() expected error after delete")
		}
	})
}

func TestClaimRepository_List(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewClaimRepository(tdb.DB)
	ctx := context.Background()

	tdb.CleanupTables(t, "user_claims", "role_claims", "claims")

	for i := 0; i < 5; i++ {
		tdb.CreateTestClaim(ctx, t)
	}

	t.Run("list with pagination", func(t *testing.T) {
		claims, err := repo.List(ctx, 0, 3)
		if err != nil {
			t.Errorf("List() error = %v", err)
			return
		}
		if len(claims) != 3 {
			t.Errorf("List() returned %d claims, want 3", len(claims))
		}
	})

	t.Run("list with offset", func(t *testing.T) {
		claims, err := repo.List(ctx, 3, 10)
		if err != nil {
			t.Errorf("List() error = %v", err)
			return
		}
		if len(claims) != 2 {
			t.Errorf("List() returned %d claims, want 2", len(claims))
		}
	})
}

func TestClaimRepository_AssignToUser(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewClaimRepository(tdb.DB)
	ctx := context.Background()

	t.Run("assign claim to user", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		claim := tdb.CreateTestClaim(ctx, t)

		err := repo.AssignToUser(ctx, user.ID, claim.ID)
		if err != nil {
			t.Errorf("AssignToUser() error = %v", err)
			return
		}

		claims, err := repo.GetUserClaims(ctx, user.ID)
		if err != nil {
			t.Errorf("GetUserClaims() error = %v", err)
			return
		}
		if len(claims) != 1 {
			t.Errorf("GetUserClaims() returned %d claims, want 1", len(claims))
		}
	})

	t.Run("assign same claim twice (idempotent)", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		claim := tdb.CreateTestClaim(ctx, t)

		_ = repo.AssignToUser(ctx, user.ID, claim.ID)
		err := repo.AssignToUser(ctx, user.ID, claim.ID)
		if err != nil {
			t.Errorf("AssignToUser() second call error = %v", err)
		}

		claims, _ := repo.GetUserClaims(ctx, user.ID)
		if len(claims) != 1 {
			t.Errorf("GetUserClaims() returned %d claims after double assign, want 1", len(claims))
		}
	})
}

func TestClaimRepository_RemoveFromUser(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewClaimRepository(tdb.DB)
	ctx := context.Background()

	t.Run("remove claim from user", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		claim := tdb.CreateTestClaim(ctx, t)

		_ = repo.AssignToUser(ctx, user.ID, claim.ID)

		err := repo.RemoveFromUser(ctx, user.ID, claim.ID)
		if err != nil {
			t.Errorf("RemoveFromUser() error = %v", err)
			return
		}

		claims, _ := repo.GetUserClaims(ctx, user.ID)
		if len(claims) != 0 {
			t.Errorf("GetUserClaims() returned %d claims after remove, want 0", len(claims))
		}
	})

	t.Run("remove non-existing assignment (no error)", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		claim := tdb.CreateTestClaim(ctx, t)

		err := repo.RemoveFromUser(ctx, user.ID, claim.ID)
		if err != nil {
			t.Errorf("RemoveFromUser() unexpected error = %v", err)
		}
	})
}

func TestClaimRepository_GetUserClaims(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewClaimRepository(tdb.DB)
	ctx := context.Background()

	t.Run("user with multiple claims", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		claim1 := tdb.CreateTestClaim(ctx, t)
		claim2 := tdb.CreateTestClaim(ctx, t)
		claim3 := tdb.CreateTestClaim(ctx, t)

		_ = repo.AssignToUser(ctx, user.ID, claim1.ID)
		_ = repo.AssignToUser(ctx, user.ID, claim2.ID)
		_ = repo.AssignToUser(ctx, user.ID, claim3.ID)

		claims, err := repo.GetUserClaims(ctx, user.ID)
		if err != nil {
			t.Errorf("GetUserClaims() error = %v", err)
			return
		}
		if len(claims) != 3 {
			t.Errorf("GetUserClaims() returned %d claims, want 3", len(claims))
		}
	})

	t.Run("user with no claims", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)

		claims, err := repo.GetUserClaims(ctx, user.ID)
		if err != nil {
			t.Errorf("GetUserClaims() error = %v", err)
			return
		}
		if len(claims) != 0 {
			t.Errorf("GetUserClaims() returned %d claims, want 0", len(claims))
		}
	})
}

func TestClaimRepository_AssignToRole(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewClaimRepository(tdb.DB)
	ctx := context.Background()

	t.Run("assign claim to role", func(t *testing.T) {
		role := tdb.CreateTestRole(ctx, t)
		claim := tdb.CreateTestClaim(ctx, t)

		err := repo.AssignToRole(ctx, role.ID, claim.ID)
		if err != nil {
			t.Errorf("AssignToRole() error = %v", err)
			return
		}

		claims, err := repo.GetRoleClaims(ctx, role.ID)
		if err != nil {
			t.Errorf("GetRoleClaims() error = %v", err)
			return
		}
		if len(claims) != 1 {
			t.Errorf("GetRoleClaims() returned %d claims, want 1", len(claims))
		}
	})

	t.Run("assign same claim twice (idempotent)", func(t *testing.T) {
		role := tdb.CreateTestRole(ctx, t)
		claim := tdb.CreateTestClaim(ctx, t)

		_ = repo.AssignToRole(ctx, role.ID, claim.ID)
		err := repo.AssignToRole(ctx, role.ID, claim.ID)
		if err != nil {
			t.Errorf("AssignToRole() second call error = %v", err)
		}

		claims, _ := repo.GetRoleClaims(ctx, role.ID)
		if len(claims) != 1 {
			t.Errorf("GetRoleClaims() returned %d claims after double assign, want 1", len(claims))
		}
	})
}

func TestClaimRepository_RemoveFromRole(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewClaimRepository(tdb.DB)
	ctx := context.Background()

	t.Run("remove claim from role", func(t *testing.T) {
		role := tdb.CreateTestRole(ctx, t)
		claim := tdb.CreateTestClaim(ctx, t)

		_ = repo.AssignToRole(ctx, role.ID, claim.ID)

		err := repo.RemoveFromRole(ctx, role.ID, claim.ID)
		if err != nil {
			t.Errorf("RemoveFromRole() error = %v", err)
			return
		}

		claims, _ := repo.GetRoleClaims(ctx, role.ID)
		if len(claims) != 0 {
			t.Errorf("GetRoleClaims() returned %d claims after remove, want 0", len(claims))
		}
	})

	t.Run("remove non-existing assignment (no error)", func(t *testing.T) {
		role := tdb.CreateTestRole(ctx, t)
		claim := tdb.CreateTestClaim(ctx, t)

		err := repo.RemoveFromRole(ctx, role.ID, claim.ID)
		if err != nil {
			t.Errorf("RemoveFromRole() unexpected error = %v", err)
		}
	})
}

func TestClaimRepository_GetRoleClaims(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewClaimRepository(tdb.DB)
	ctx := context.Background()

	t.Run("role with multiple claims", func(t *testing.T) {
		role := tdb.CreateTestRole(ctx, t)
		claim1 := tdb.CreateTestClaim(ctx, t)
		claim2 := tdb.CreateTestClaim(ctx, t)

		_ = repo.AssignToRole(ctx, role.ID, claim1.ID)
		_ = repo.AssignToRole(ctx, role.ID, claim2.ID)

		claims, err := repo.GetRoleClaims(ctx, role.ID)
		if err != nil {
			t.Errorf("GetRoleClaims() error = %v", err)
			return
		}
		if len(claims) != 2 {
			t.Errorf("GetRoleClaims() returned %d claims, want 2", len(claims))
		}
	})

	t.Run("role with no claims", func(t *testing.T) {
		role := tdb.CreateTestRole(ctx, t)

		claims, err := repo.GetRoleClaims(ctx, role.ID)
		if err != nil {
			t.Errorf("GetRoleClaims() error = %v", err)
			return
		}
		if len(claims) != 0 {
			t.Errorf("GetRoleClaims() returned %d claims, want 0", len(claims))
		}
	})
}

var _ entity.Claim
