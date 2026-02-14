package repository

import (
	"context"
	"testing"

	"github.com/creafly/identity/internal/testutil"
	"github.com/creafly/identity/internal/utils"
)

func TestUserRepository_Create(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)
	ctx := context.Background()

	t.Run("valid user", func(t *testing.T) {
		user := testutil.NewTestUser()

		err := repo.Create(ctx, user)
		if err != nil {
			t.Errorf("Create() error = %v", err)
			return
		}

		got, err := repo.GetByID(ctx, user.ID)
		if err != nil {
			t.Errorf("GetByID() error = %v", err)
			return
		}
		if got.Email != user.Email {
			t.Errorf("Create() email = %v, want %v", got.Email, user.Email)
		}
	})

	t.Run("duplicate email should fail", func(t *testing.T) {
		user1 := testutil.NewTestUser()
		_ = repo.Create(ctx, user1)

		user2 := testutil.NewTestUser()
		user2.Email = user1.Email

		err := repo.Create(ctx, user2)
		if err == nil {
			t.Error("Create() expected error for duplicate email")
		}
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)
	ctx := context.Background()

	t.Run("existing user", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)

		got, err := repo.GetByID(ctx, user.ID)
		if err != nil {
			t.Errorf("GetByID() error = %v", err)
			return
		}
		if got.ID != user.ID {
			t.Errorf("GetByID() id = %v, want %v", got.ID, user.ID)
		}
	})

	t.Run("non-existing user", func(t *testing.T) {
		_, err := repo.GetByID(ctx, utils.GenerateUUID())
		if err == nil {
			t.Error("GetByID() expected error for non-existing user")
		}
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)
	ctx := context.Background()

	t.Run("existing email", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)

		got, err := repo.GetByEmail(ctx, user.Email)
		if err != nil {
			t.Errorf("GetByEmail() error = %v", err)
			return
		}
		if got.Email != user.Email {
			t.Errorf("GetByEmail() email = %v, want %v", got.Email, user.Email)
		}
	})

	t.Run("non-existing email", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		if err == nil {
			t.Error("GetByEmail() expected error for non-existing email")
		}
	})
}

func TestUserRepository_Update(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)
	ctx := context.Background()

	t.Run("update user", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)

		user.FirstName = "UpdatedFirst"
		user.LastName = "UpdatedLast"
		err := repo.Update(ctx, user)
		if err != nil {
			t.Errorf("Update() error = %v", err)
			return
		}

		got, _ := repo.GetByID(ctx, user.ID)
		if got.FirstName != "UpdatedFirst" {
			t.Errorf("Update() firstName = %v, want UpdatedFirst", got.FirstName)
		}
	})
}

func TestUserRepository_Delete(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)
	ctx := context.Background()

	t.Run("soft delete user", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)

		err := repo.Delete(ctx, user.ID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
			return
		}

		_, err = repo.GetByID(ctx, user.ID)
		if err == nil {
			t.Error("Delete() user still accessible after soft delete")
		}
	})
}

func TestUserRepository_List(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)
	ctx := context.Background()

	t.Run("list users", func(t *testing.T) {
		_ = tdb.CreateTestUser(ctx, t)
		_ = tdb.CreateTestUser(ctx, t)

		users, err := repo.List(ctx, 0, 10)
		if err != nil {
			t.Errorf("List() error = %v", err)
			return
		}
		if len(users) < 2 {
			t.Errorf("List() count = %d, want >= 2", len(users))
		}
	})
}

func TestUserRepository_UpdateBlocked(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)
	ctx := context.Background()

	t.Run("block user", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		reason := "Test block reason"

		err := repo.UpdateBlocked(ctx, user.ID, true, &reason, nil)
		if err != nil {
			t.Errorf("UpdateBlocked() error = %v", err)
			return
		}

		got, _ := repo.GetByID(ctx, user.ID)
		if !got.IsBlocked {
			t.Error("UpdateBlocked() user not blocked")
		}
	})

	t.Run("unblock user", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		reason := "Test block reason"
		_ = repo.UpdateBlocked(ctx, user.ID, true, &reason, nil)

		err := repo.UpdateBlocked(ctx, user.ID, false, nil, nil)
		if err != nil {
			t.Errorf("UpdateBlocked() error = %v", err)
			return
		}

		got, _ := repo.GetByID(ctx, user.ID)
		if got.IsBlocked {
			t.Error("UpdateBlocked() user still blocked after unblock")
		}
	})
}

func TestUserRepository_TOTP(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)
	ctx := context.Background()

	t.Run("set totp secret", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)

		err := repo.SetTOTPSecret(ctx, user.ID, "test-secret-key")
		if err != nil {
			t.Errorf("SetTOTPSecret() error = %v", err)
			return
		}

		got, _ := repo.GetByID(ctx, user.ID)
		if got.TotpSecret == nil || *got.TotpSecret != "test-secret-key" {
			t.Error("SetTOTPSecret() secret not set correctly")
		}
	})

	t.Run("enable totp", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		_ = repo.SetTOTPSecret(ctx, user.ID, "test-secret")

		err := repo.EnableTOTP(ctx, user.ID)
		if err != nil {
			t.Errorf("EnableTOTP() error = %v", err)
			return
		}

		got, _ := repo.GetByID(ctx, user.ID)
		if !got.TotpEnabled {
			t.Error("EnableTOTP() totp not enabled")
		}
	})

	t.Run("disable totp", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		_ = repo.SetTOTPSecret(ctx, user.ID, "test-secret")
		_ = repo.EnableTOTP(ctx, user.ID)

		err := repo.DisableTOTP(ctx, user.ID)
		if err != nil {
			t.Errorf("DisableTOTP() error = %v", err)
			return
		}

		got, _ := repo.GetByID(ctx, user.ID)
		if got.TotpEnabled {
			t.Error("DisableTOTP() totp still enabled")
		}
	})
}
