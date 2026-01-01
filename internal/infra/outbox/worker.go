package outbox

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/hexaend/identity/internal/domain/entity"
	"github.com/hexaend/identity/internal/domain/repository"
)

func calculateRetryDelay(retryCount int) time.Duration {
	switch {
	case retryCount < 3:
		return time.Duration(1<<retryCount) * time.Second
	case retryCount < 5:
		delays := []time.Duration{30 * time.Second, time.Minute}
		return delays[retryCount-3]
	default:
		delays := []time.Duration{
			5 * time.Minute,
			10 * time.Minute,
			30 * time.Minute,
			time.Hour,
			2 * time.Hour,
		}
		idx := retryCount - 5
		if idx >= len(delays) {
			return delays[len(delays)-1]
		}
		return delays[idx]
	}
}

type WorkerConfig struct {
	PollInterval          time.Duration
	BatchSize             int
	CleanupInterval       time.Duration
	RetentionPeriod       time.Duration
	FailedRetentionPeriod time.Duration
}

func DefaultConfig() WorkerConfig {
	return WorkerConfig{
		PollInterval:          time.Second,
		BatchSize:             100,
		CleanupInterval:       time.Hour,
		RetentionPeriod:       24 * time.Hour,
		FailedRetentionPeriod: 7 * 24 * time.Hour,
	}
}

type EventHandler interface {
	Handle(ctx context.Context, event *entity.OutboxEvent) error
}

type LogEventHandler struct{}

func NewLogEventHandler() *LogEventHandler {
	return &LogEventHandler{}
}

func (h *LogEventHandler) Handle(ctx context.Context, event *entity.OutboxEvent) error {
	log.Printf("[Outbox LogHandler] Event: type=%s, id=%s, payload=%s",
		event.EventType, event.ID, event.Payload)
	return nil
}

type KafkaEventHandler struct {
	producer sarama.SyncProducer
	topicMap map[string]string
}

func NewKafkaEventHandler(producer sarama.SyncProducer) *KafkaEventHandler {
	return &KafkaEventHandler{
		producer: producer,
		topicMap: map[string]string{
			"users.registered":       "users",
			"users.updated":          "users",
			"users.blocked":          "users",
			"users.unblocked":        "users",
			"users.password_changed": "users",
			"tenants.created":        "tenants",
			"tenants.updated":        "tenants",
			"tenants.deleted":        "tenants",
			"tenants.member_added":   "tenants",
			"tenants.member_removed": "tenants",
			"invitations.requested":  "invitations",
		},
	}
}

func (h *KafkaEventHandler) Handle(ctx context.Context, event *entity.OutboxEvent) error {
	topic := h.getTopic(event.EventType)

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(event.ID.String()),
		Value: sarama.StringEncoder(event.Payload),
		Headers: []sarama.RecordHeader{
			{Key: []byte("event_type"), Value: []byte(event.EventType)},
			{Key: []byte("event_id"), Value: []byte(event.ID.String())},
			{Key: []byte("created_at"), Value: []byte(event.CreatedAt.Format(time.RFC3339))},
		},
	}

	_, _, err := h.producer.SendMessage(msg)
	return err
}

func (h *KafkaEventHandler) getTopic(eventType string) string {
	if topic, ok := h.topicMap[eventType]; ok {
		return topic
	}
	return "events"
}

type Worker struct {
	repo    repository.OutboxRepository
	handler EventHandler
	config  WorkerConfig
	wg      sync.WaitGroup
	stopCh  chan struct{}
}

func NewWorker(repo repository.OutboxRepository, handler EventHandler, config WorkerConfig) *Worker {
	return &Worker{
		repo:    repo,
		handler: handler,
		config:  config,
		stopCh:  make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("[Outbox Worker] Starting...")

	w.wg.Add(1)
	go w.processLoop(ctx)

	w.wg.Add(1)
	go w.cleanupLoop(ctx)

	log.Println("[Outbox Worker] Started")
}

func (w *Worker) Stop() {
	log.Println("[Outbox Worker] Stopping...")
	close(w.stopCh)
	w.wg.Wait()
	log.Println("[Outbox Worker] Stopped")
}

func (w *Worker) processLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	events, err := w.repo.GetPending(ctx, w.config.BatchSize)
	if err != nil {
		log.Printf("[Outbox Worker] Error fetching pending events: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	log.Printf("[Outbox Worker] Processing %d events", len(events))

	for _, event := range events {
		if err := w.processEvent(ctx, event); err != nil {
			log.Printf("[Outbox Worker] Error processing event %s: %v", event.ID, err)
		}
	}
}

func (w *Worker) processEvent(ctx context.Context, event *entity.OutboxEvent) error {
	err := w.handler.Handle(ctx, event)
	if err != nil {
		if event.RetryCount >= repository.MaxRetryCount-1 {
			log.Printf("[Outbox Worker] Event %s failed after max retries, marking as failed", event.ID)
			return w.repo.MarkAsFailed(ctx, event.ID)
		}

		delay := calculateRetryDelay(event.RetryCount)
		nextRetryAt := time.Now().Add(delay)

		log.Printf("[Outbox Worker] Event %s failed (retry %d), next retry at %s",
			event.ID, event.RetryCount+1, nextRetryAt.Format(time.RFC3339))

		return w.repo.IncrementRetry(ctx, event.ID, nextRetryAt)
	}

	return w.repo.MarkProcessed(ctx, event.ID)
}

func (w *Worker) cleanupLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runCleanup(ctx)
		}
	}
}

func (w *Worker) runCleanup(ctx context.Context) {
	log.Println("[Outbox Worker] Running cleanup...")

	if err := w.repo.DeleteOldProcessed(ctx, w.config.RetentionPeriod); err != nil {
		log.Printf("[Outbox Worker] Error cleaning up old events: %v", err)
	}

	if err := w.repo.DeleteOldFailed(ctx, w.config.FailedRetentionPeriod); err != nil {
		log.Printf("[Outbox Worker] Error cleaning up failed events: %v", err)
	}

	log.Println("[Outbox Worker] Cleanup completed")
}

func CreatePayload(data any) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
