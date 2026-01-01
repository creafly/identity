package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hexaend/identity/internal/domain/entity"
	"github.com/hexaend/identity/internal/domain/repository"
	"github.com/hexaend/identity/internal/infra/outbox"
)

type InvitationService interface {
	RequestInvitation(ctx context.Context, input RequestInvitationInput) error
}

type RequestInvitationInput struct {
	TenantID    uuid.UUID
	TenantName  string
	InviterID   uuid.UUID
	InviterName string
	InviteeID   uuid.UUID
	Email       string
}

type invitationService struct {
	outboxRepo repository.OutboxRepository
}

func NewInvitationService(outboxRepo repository.OutboxRepository) InvitationService {
	return &invitationService{outboxRepo: outboxRepo}
}

func (s *invitationService) RequestInvitation(ctx context.Context, input RequestInvitationInput) error {
	payload, err := outbox.CreatePayload(map[string]any{
		"tenantId":    input.TenantID,
		"tenantName":  input.TenantName,
		"inviterId":   input.InviterID,
		"inviterName": input.InviterName,
		"inviteeId":   input.InviteeID,
		"email":       input.Email,
	})
	if err != nil {
		return err
	}

	event := &entity.OutboxEvent{
		ID:        uuid.New(),
		EventType: "invitations.requested",
		Payload:   payload,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	return s.outboxRepo.Create(ctx, event)
}
