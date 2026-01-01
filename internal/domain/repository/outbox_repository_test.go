package repository

import (
	"context"
	"testing"
	"time"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/testutil"
)

func TestOutboxRepository_Create(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	t.Run("valid event", func(t *testing.T) {
		event := testutil.NewTestOutboxEvent()
		err := repo.Create(ctx, event)

		if err != nil {
			t.Errorf("Create() error = %v", err)
		}
	})
}

func TestOutboxRepository_GetPending(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	tdb.CleanupTables(t, "outbox_events")

	t.Run("returns pending events", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			event := testutil.NewTestOutboxEvent()
			_ = repo.Create(ctx, event)
		}

		events, err := repo.GetPending(ctx, 10)
		if err != nil {
			t.Errorf("GetPending() error = %v", err)
			return
		}
		if len(events) != 3 {
			t.Errorf("GetPending() returned %d events, want 3", len(events))
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		events, err := repo.GetPending(ctx, 2)
		if err != nil {
			t.Errorf("GetPending() error = %v", err)
			return
		}
		if len(events) > 2 {
			t.Errorf("GetPending() returned %d events, want <= 2", len(events))
		}
	})

	t.Run("excludes processed events", func(t *testing.T) {
		tdb.CleanupTables(t, "outbox_events")

		event := testutil.NewTestOutboxEvent()
		_ = repo.Create(ctx, event)
		_ = repo.MarkProcessed(ctx, event.ID)

		events, err := repo.GetPending(ctx, 10)
		if err != nil {
			t.Errorf("GetPending() error = %v", err)
			return
		}
		if len(events) != 0 {
			t.Errorf("GetPending() returned %d events, want 0", len(events))
		}
	})
}

func TestOutboxRepository_MarkProcessed(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	t.Run("mark event as processed", func(t *testing.T) {
		event := testutil.NewTestOutboxEvent()
		_ = repo.Create(ctx, event)

		err := repo.MarkProcessed(ctx, event.ID)
		if err != nil {
			t.Errorf("MarkProcessed() error = %v", err)
			return
		}

		events, _ := repo.GetPending(ctx, 10)
		for _, e := range events {
			if e.ID == event.ID {
				t.Error("MarkProcessed() event still in pending list")
			}
		}
	})
}

func TestOutboxRepository_MarkAsFailed(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	t.Run("mark event as failed", func(t *testing.T) {
		event := testutil.NewTestOutboxEvent()
		_ = repo.Create(ctx, event)

		err := repo.MarkAsFailed(ctx, event.ID)
		if err != nil {
			t.Errorf("MarkAsFailed() error = %v", err)
			return
		}

		events, _ := repo.GetPending(ctx, 10)
		for _, e := range events {
			if e.ID == event.ID {
				t.Error("MarkAsFailed() event still in pending list")
			}
		}
	})
}

func TestOutboxRepository_IncrementRetry(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	t.Run("increment retry count", func(t *testing.T) {
		tdb.CleanupTables(t, "outbox_events")

		event := testutil.NewTestOutboxEvent()
		_ = repo.Create(ctx, event)

		nextRetry := time.Now().Add(1 * time.Hour)
		err := repo.IncrementRetry(ctx, event.ID, nextRetry)
		if err != nil {
			t.Errorf("IncrementRetry() error = %v", err)
			return
		}

		events, _ := repo.GetPending(ctx, 10)
		if len(events) != 0 {
			t.Errorf("GetPending() returned %d events, want 0", len(events))
		}
	})
}

var _ entity.OutboxEvent
