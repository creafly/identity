package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotificationsServiceUnavailable = errors.New("notifications service unavailable")
	ErrInvitationFailed                = errors.New("failed to create invitation")
	ErrInvitationNotFound              = errors.New("invitation not found")
)

type NotificationsClient interface {
	CreateInvitation(ctx context.Context, input CreateInvitationInput) (*Invitation, error)
	CancelInvitation(ctx context.Context, invitationID uuid.UUID) error
}

type CreateInvitationInput struct {
	TenantID   uuid.UUID `json:"tenantId"`
	TenantName string    `json:"tenantName"`
	InviterID  uuid.UUID `json:"inviterId"`
	InviteeID  uuid.UUID `json:"inviteeId"`
}

type Invitation struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	InviterID uuid.UUID `json:"inviterId"`
	InviteeID uuid.UUID `json:"inviteeId"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type notificationsClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewNotificationsClient(baseURL string) NotificationsClient {
	return &notificationsClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *notificationsClient) CreateInvitation(ctx context.Context, input CreateInvitationInput) (*Invitation, error) {
	url := fmt.Sprintf("%s/api/v1/invitations", c.baseURL)

	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Name", "identity")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, ErrNotificationsServiceUnavailable
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 500 {
		return nil, ErrNotificationsServiceUnavailable
	}

	if resp.StatusCode >= 400 {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, ErrInvitationFailed
		}
		return nil, fmt.Errorf("%w: %s", ErrInvitationFailed, errorResp.Error)
	}

	var response struct {
		Invitation *Invitation `json:"invitation"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Invitation, nil
}

func (c *notificationsClient) CancelInvitation(ctx context.Context, invitationID uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/invitations/%s/cancel", c.baseURL, invitationID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Name", "identity")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ErrNotificationsServiceUnavailable
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return ErrInvitationNotFound
	}

	if resp.StatusCode >= 400 {
		return ErrInvitationFailed
	}

	return nil
}
