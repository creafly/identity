package mocks

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/creafly/identity/internal/domain/entity"
)

type ClaimRepositoryMock struct {
	mu          sync.RWMutex
	claims      map[uuid.UUID]*entity.Claim
	userClaims  map[uuid.UUID][]uuid.UUID
	roleClaims  map[uuid.UUID][]uuid.UUID
	CreateFunc  func(ctx context.Context, claim *entity.Claim) error
	GetByIDFunc func(ctx context.Context, id uuid.UUID) (*entity.Claim, error)
	DeleteFunc  func(ctx context.Context, id uuid.UUID) error
}

func NewClaimRepositoryMock() *ClaimRepositoryMock {
	return &ClaimRepositoryMock{
		claims:     make(map[uuid.UUID]*entity.Claim),
		userClaims: make(map[uuid.UUID][]uuid.UUID),
		roleClaims: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *ClaimRepositoryMock) Create(ctx context.Context, claim *entity.Claim) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, claim)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range m.claims {
		if c.Value == claim.Value {
			return sql.ErrNoRows
		}
	}

	m.claims[claim.ID] = claim
	return nil
}

func (m *ClaimRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Claim, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	claim, ok := m.claims[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return claim, nil
}

func (m *ClaimRepositoryMock) GetByValue(ctx context.Context, value string) (*entity.Claim, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, claim := range m.claims {
		if claim.Value == value {
			return claim, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *ClaimRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.claims, id)
	return nil
}

func (m *ClaimRepositoryMock) List(ctx context.Context, offset, limit int) ([]*entity.Claim, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var claims []*entity.Claim
	for _, claim := range m.claims {
		claims = append(claims, claim)
	}

	if offset >= len(claims) {
		return []*entity.Claim{}, nil
	}
	end := offset + limit
	if end > len(claims) {
		end = len(claims)
	}
	return claims[offset:end], nil
}

func (m *ClaimRepositoryMock) AssignToUser(ctx context.Context, userID, claimID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.claims[claimID]; !ok {
		return sql.ErrNoRows
	}

	m.userClaims[userID] = append(m.userClaims[userID], claimID)
	return nil
}

func (m *ClaimRepositoryMock) RemoveFromUser(ctx context.Context, userID, claimID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	claims := m.userClaims[userID]
	for i, id := range claims {
		if id == claimID {
			m.userClaims[userID] = append(claims[:i], claims[i+1:]...)
			break
		}
	}
	return nil
}

func (m *ClaimRepositoryMock) GetUserClaims(ctx context.Context, userID uuid.UUID) ([]*entity.Claim, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*entity.Claim
	for _, claimID := range m.userClaims[userID] {
		if claim, ok := m.claims[claimID]; ok {
			result = append(result, claim)
		}
	}
	return result, nil
}

func (m *ClaimRepositoryMock) AssignToRole(ctx context.Context, roleID, claimID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.claims[claimID]; !ok {
		return sql.ErrNoRows
	}

	m.roleClaims[roleID] = append(m.roleClaims[roleID], claimID)
	return nil
}

func (m *ClaimRepositoryMock) RemoveFromRole(ctx context.Context, roleID, claimID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	claims := m.roleClaims[roleID]
	for i, id := range claims {
		if id == claimID {
			m.roleClaims[roleID] = append(claims[:i], claims[i+1:]...)
			break
		}
	}
	return nil
}

func (m *ClaimRepositoryMock) GetRoleClaims(ctx context.Context, roleID uuid.UUID) ([]*entity.Claim, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*entity.Claim
	for _, claimID := range m.roleClaims[roleID] {
		if claim, ok := m.claims[claimID]; ok {
			result = append(result, claim)
		}
	}
	return result, nil
}

func (m *ClaimRepositoryMock) AddClaim(claim *entity.Claim) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.claims[claim.ID] = claim
}

type OutboxRepositoryMock struct {
	mu     sync.RWMutex
	events map[uuid.UUID]*entity.OutboxEvent
}

func NewOutboxRepositoryMock() *OutboxRepositoryMock {
	return &OutboxRepositoryMock{
		events: make(map[uuid.UUID]*entity.OutboxEvent),
	}
}

func (m *OutboxRepositoryMock) Create(ctx context.Context, event *entity.OutboxEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if event.ID == uuid.Nil {
		return sql.ErrNoRows
	}

	m.events[event.ID] = event
	return nil
}

func (m *OutboxRepositoryMock) GetPending(ctx context.Context, limit int) ([]*entity.OutboxEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*entity.OutboxEvent
	for _, event := range m.events {
		if event.Status == "pending" && event.RetryCount < 10 {
			if event.NextRetryAt == nil || event.NextRetryAt.Before(time.Now()) {
				result = append(result, event)
				if len(result) >= limit {
					break
				}
			}
		}
	}
	return result, nil
}

func (m *OutboxRepositoryMock) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	event, ok := m.events[id]
	if !ok {
		return sql.ErrNoRows
	}
	event.Status = "processed"
	now := time.Now()
	event.ProcessedAt = &now
	return nil
}

func (m *OutboxRepositoryMock) MarkAsFailed(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	event, ok := m.events[id]
	if !ok {
		return sql.ErrNoRows
	}
	event.Status = "failed"
	return nil
}

func (m *OutboxRepositoryMock) IncrementRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	event, ok := m.events[id]
	if !ok {
		return sql.ErrNoRows
	}
	event.RetryCount++
	event.NextRetryAt = &nextRetryAt
	now := time.Now()
	event.LastErrorAt = &now
	return nil
}

func (m *OutboxRepositoryMock) DeleteOldProcessed(ctx context.Context, olderThan time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	for id, event := range m.events {
		if event.Status == "processed" && event.ProcessedAt != nil && event.ProcessedAt.Before(cutoff) {
			delete(m.events, id)
		}
	}
	return nil
}

func (m *OutboxRepositoryMock) DeleteOldFailed(ctx context.Context, olderThan time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	for id, event := range m.events {
		if event.Status == "failed" && event.CreatedAt.Before(cutoff) {
			delete(m.events, id)
		}
	}
	return nil
}

func (m *OutboxRepositoryMock) AddEvent(event *entity.OutboxEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[event.ID] = event
}

func (m *OutboxRepositoryMock) GetEvent(id uuid.UUID) *entity.OutboxEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.events[id]
}

type PasswordResetRepositoryMock struct {
	mu     sync.RWMutex
	tokens map[uuid.UUID]*entity.PasswordResetToken
}

func NewPasswordResetRepositoryMock() *PasswordResetRepositoryMock {
	return &PasswordResetRepositoryMock{
		tokens: make(map[uuid.UUID]*entity.PasswordResetToken),
	}
}

func (m *PasswordResetRepositoryMock) Create(ctx context.Context, token *entity.PasswordResetToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if token.ID == uuid.Nil || token.UserID == uuid.Nil {
		return sql.ErrNoRows
	}

	m.tokens[token.ID] = token
	return nil
}

func (m *PasswordResetRepositoryMock) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.PasswordResetToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, token := range m.tokens {
		if token.TokenHash == tokenHash && token.UsedAt == nil && token.ExpiresAt.After(time.Now()) {
			return token, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *PasswordResetRepositoryMock) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	token, ok := m.tokens[id]
	if !ok {
		return sql.ErrNoRows
	}
	now := time.Now()
	token.UsedAt = &now
	return nil
}

func (m *PasswordResetRepositoryMock) DeleteExpiredTokens(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, token := range m.tokens {
		if token.ExpiresAt.Before(time.Now()) || token.UsedAt != nil {
			delete(m.tokens, id)
		}
	}
	return nil
}

func (m *PasswordResetRepositoryMock) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, token := range m.tokens {
		if token.UserID == userID {
			delete(m.tokens, id)
		}
	}
	return nil
}

func (m *PasswordResetRepositoryMock) AddToken(token *entity.PasswordResetToken) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[token.ID] = token
}

func (m *PasswordResetRepositoryMock) GetToken(id uuid.UUID) *entity.PasswordResetToken {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tokens[id]
}

func (m *PasswordResetRepositoryMock) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tokens)
}

func (m *PasswordResetRepositoryMock) CountByUserID(userID uuid.UUID) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, token := range m.tokens {
		if token.UserID == userID {
			count++
		}
	}
	return count
}
