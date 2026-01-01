package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/creafly/identity/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

const MaxRetryCount = 10

type OutboxRepository interface {
	Create(ctx context.Context, event *entity.OutboxEvent) error
	GetPending(ctx context.Context, limit int) ([]*entity.OutboxEvent, error)
	MarkProcessed(ctx context.Context, id uuid.UUID) error
	MarkAsFailed(ctx context.Context, id uuid.UUID) error
	IncrementRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error
	DeleteOldProcessed(ctx context.Context, olderThan time.Duration) error
	DeleteOldFailed(ctx context.Context, olderThan time.Duration) error
}

type outboxRepository struct {
	db *sqlx.DB
}

func NewOutboxRepository(db *sqlx.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) Create(ctx context.Context, event *entity.OutboxEvent) error {
	query := `
		INSERT INTO outbox_events (id, event_type, payload, status, retry_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		event.ID,
		event.EventType,
		event.Payload,
		event.Status,
		event.RetryCount,
		event.CreatedAt,
	)
	return err
}

func (r *outboxRepository) GetPending(ctx context.Context, limit int) ([]*entity.OutboxEvent, error) {
	var events []*entity.OutboxEvent
	query := `
		SELECT id, event_type, payload, status, retry_count, next_retry_at, last_error_at, processed_at, created_at 
		FROM outbox_events 
		WHERE status = 'pending' 
		  AND retry_count < $1
		  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY created_at ASC 
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &events, query, MaxRetryCount, limit)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *outboxRepository) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE outbox_events SET status = 'processed', processed_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *outboxRepository) MarkAsFailed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE outbox_events SET status = 'failed' WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *outboxRepository) IncrementRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error {
	query := `
		UPDATE outbox_events 
		SET retry_count = retry_count + 1, 
		    next_retry_at = $1,
		    last_error_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, nextRetryAt, id)
	return err
}

func (r *outboxRepository) DeleteOldProcessed(ctx context.Context, olderThan time.Duration) error {
	query := `DELETE FROM outbox_events WHERE status = 'processed' AND processed_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now().Add(-olderThan))
	return err
}

func (r *outboxRepository) DeleteOldFailed(ctx context.Context, olderThan time.Duration) error {
	query := `DELETE FROM outbox_events WHERE status = 'failed' AND created_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now().Add(-olderThan))
	return err
}
