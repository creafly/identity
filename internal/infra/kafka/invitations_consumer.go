package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/creafly/identity/internal/domain/service"
)

type InvitationAcceptedEvent struct {
	InvitationID uuid.UUID  `json:"invitationId"`
	TenantID     uuid.UUID  `json:"tenantId"`
	TenantName   string     `json:"tenantName"`
	InviteeID    uuid.UUID  `json:"inviteeId"`
	Email        string     `json:"email"`
	RoleID       *uuid.UUID `json:"roleId"`
	InviterID    uuid.UUID  `json:"inviterId"`
}

type InvitationsConsumer struct {
	tenantService service.TenantService
	consumer      sarama.ConsumerGroup
	topics        []string
	ready         chan bool
}

func NewInvitationsConsumer(
	brokers []string,
	groupID string,
	tenantService service.TenantService,
) (*InvitationsConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &InvitationsConsumer{
		tenantService: tenantService,
		consumer:      consumer,
		topics:        []string{"invitations"},
		ready:         make(chan bool),
	}, nil
}

func (c *InvitationsConsumer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := c.consumer.Consume(ctx, c.topics, c); err != nil {
					log.Printf("[InvitationsConsumer] Error consuming: %v", err)
				}
				if ctx.Err() != nil {
					return
				}
				c.ready = make(chan bool)
			}
		}
	}()

	<-c.ready
	log.Println("[InvitationsConsumer] Started and ready")
}

func (c *InvitationsConsumer) Stop() error {
	return c.consumer.Close()
}

func (c *InvitationsConsumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *InvitationsConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *InvitationsConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Println("[InvitationsConsumer] Message channel was closed")
				return nil
			}

			if err := c.processMessage(session.Context(), message); err != nil {
				log.Printf("[InvitationsConsumer] Error processing message: %v", err)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (c *InvitationsConsumer) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	var eventType string
	for _, header := range message.Headers {
		if string(header.Key) == "event_type" {
			eventType = string(header.Value)
			break
		}
	}

	log.Printf("[InvitationsConsumer] Received event: type=%s, key=%s", eventType, string(message.Key))

	switch eventType {
	case "invitations.accepted":
		return c.handleInvitationAccepted(ctx, message.Value)
	default:
		log.Printf("[InvitationsConsumer] Unknown event type: %s", eventType)
		return nil
	}
}

func (c *InvitationsConsumer) handleInvitationAccepted(ctx context.Context, value []byte) error {
	var event InvitationAcceptedEvent
	if err := json.Unmarshal(value, &event); err != nil {
		log.Printf("[InvitationsConsumer] Error unmarshaling event: %v", err)
		return err
	}

	log.Printf("[InvitationsConsumer] Adding member %s to tenant %s", event.InviteeID, event.TenantID)

	if err := c.tenantService.AddMember(ctx, event.TenantID, event.InviteeID); err != nil {
		log.Printf("[InvitationsConsumer] Error adding member: %v", err)
		return err
	}

	log.Printf("[InvitationsConsumer] Member %s added to tenant %s successfully", event.InviteeID, event.TenantID)
	return nil
}
