package repository

import (
	"context"
	"testing"
	"time"

	"github.com/hexaend/identity/internal/domain/entity"
	"github.com/hexaend/identity/internal/testutil"
)

func TestPasswordResetRepository_Create(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewPasswordResetRepository(tdb.DB)
	ctx := context.Background()

	t.Run("valid token", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		token := testutil.NewTestPasswordResetToken(user.ID)

		err := repo.Create(ctx, token)
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}
	})
}

func TestPasswordResetRepository_GetByTokenHash(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewPasswordResetRepository(tdb.DB)
	ctx := context.Background()

	t.Run("valid token hash", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		token := testutil.NewTestPasswordResetToken(user.ID)
		_ = repo.Create(ctx, token)

		got, err := repo.GetByTokenHash(ctx, token.TokenHash)
		if err != nil {
			t.Errorf("GetByTokenHash() error = %v", err)
			return
		}
		if got.ID != token.ID {
			t.Errorf("GetByTokenHash() ID = %v, want %v", got.ID, token.ID)
		}
	})

	t.Run("non-existing hash", func(t *testing.T) {
		_, err := repo.GetByTokenHash(ctx, "non-existing-hash")
		if err == nil {
			t.Error("GetByTokenHash() expected error for non-existing hash")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		token := testutil.NewTestPasswordResetToken(user.ID)
		token.ExpiresAt = time.Now().Add(-1 * time.Hour)
		_ = repo.Create(ctx, token)

		_, err := repo.GetByTokenHash(ctx, token.TokenHash)
		if err == nil {
			t.Error("GetByTokenHash() expected error for expired token")
		}
	})
}

func TestPasswordResetRepository_MarkAsUsed(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewPasswordResetRepository(tdb.DB)
	ctx := context.Background()

	t.Run("mark token as used", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)
		token := testutil.NewTestPasswordResetToken(user.ID)
		_ = repo.Create(ctx, token)

		err := repo.MarkAsUsed(ctx, token.ID)
		if err != nil {
			t.Errorf("MarkAsUsed() error = %v", err)
			return
		}

		_, err = repo.GetByTokenHash(ctx, token.TokenHash)
		if err == nil {
			t.Error("GetByTokenHash() expected error for used token")
		}
	})
}

func TestPasswordResetRepository_DeleteExpiredTokens(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewPasswordResetRepository(tdb.DB)
	ctx := context.Background()

	tdb.CleanupTables(t, "password_reset_tokens")

	t.Run("delete expired tokens", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)

		expiredToken := testutil.NewTestPasswordResetToken(user.ID)
		expiredToken.ExpiresAt = time.Now().Add(-1 * time.Hour)
		_ = repo.Create(ctx, expiredToken)

		validToken := testutil.NewTestPasswordResetToken(user.ID)
		_ = repo.Create(ctx, validToken)

		err := repo.DeleteExpiredTokens(ctx)
		if err != nil {
			t.Errorf("DeleteExpiredTokens() error = %v", err)
			return
		}

		got, err := repo.GetByTokenHash(ctx, validToken.TokenHash)
		if err != nil {
			t.Errorf("GetByTokenHash() valid token error = %v", err)
			return
		}
		if got.ID != validToken.ID {
			t.Error("Valid token was deleted")
		}
	})
}

func TestPasswordResetRepository_DeleteByUserID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewPasswordResetRepository(tdb.DB)
	ctx := context.Background()

	t.Run("delete all user tokens", func(t *testing.T) {
		user := tdb.CreateTestUser(ctx, t)

		for i := 0; i < 3; i++ {
			token := testutil.NewTestPasswordResetToken(user.ID)
			_ = repo.Create(ctx, token)
		}

		err := repo.DeleteByUserID(ctx, user.ID)
		if err != nil {
			t.Errorf("DeleteByUserID() error = %v", err)
		}
	})
}

var _ entity.PasswordResetToken
