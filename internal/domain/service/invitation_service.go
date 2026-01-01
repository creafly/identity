package service

import (
	"context"

	"github.com/creafly/outbox"
	"github.com/google/uuid"
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
	outboxRepo outbox.Repository
}

func NewInvitationService(outboxRepo outbox.Repository) InvitationService {
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

	event := outbox.NewEvent("invitations.requested", payload)

	return s.outboxRepo.Create(ctx, event)
}
