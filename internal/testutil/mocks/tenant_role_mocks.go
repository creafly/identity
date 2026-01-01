package mocks

import (
	"context"
	"database/sql"
	"sync"

	"github.com/google/uuid"
	"github.com/creafly/identity/internal/domain/entity"
)

type TenantRoleRepositoryMock struct {
	mu              sync.RWMutex
	roles           map[uuid.UUID]*entity.TenantRole
	roleClaims      map[uuid.UUID][]uuid.UUID
	userTenantRoles map[string][]uuid.UUID
	claims          map[uuid.UUID]*entity.Claim
}

func NewTenantRoleRepositoryMock() *TenantRoleRepositoryMock {
	return &TenantRoleRepositoryMock{
		roles:           make(map[uuid.UUID]*entity.TenantRole),
		roleClaims:      make(map[uuid.UUID][]uuid.UUID),
		userTenantRoles: make(map[string][]uuid.UUID),
		claims:          make(map[uuid.UUID]*entity.Claim),
	}
}

func (m *TenantRoleRepositoryMock) Create(ctx context.Context, role *entity.TenantRole) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
	return nil
}

func (m *TenantRoleRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.TenantRole, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	role, ok := m.roles[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return role, nil
}

func (m *TenantRoleRepositoryMock) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*entity.TenantRole, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, role := range m.roles {
		if role.TenantID == tenantID && role.Name == name {
			return role, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *TenantRoleRepositoryMock) Update(ctx context.Context, role *entity.TenantRole) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
	return nil
}

func (m *TenantRoleRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.roles, id)
	return nil
}

func (m *TenantRoleRepositoryMock) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantRole, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entity.TenantRole
	for _, role := range m.roles {
		if role.TenantID == tenantID {
			result = append(result, role)
		}
	}
	return result, nil
}

func (m *TenantRoleRepositoryMock) GetDefaultRoles(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantRole, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entity.TenantRole
	for _, role := range m.roles {
		if role.TenantID == tenantID && role.IsDefault {
			result = append(result, role)
		}
	}
	return result, nil
}

func (m *TenantRoleRepositoryMock) AddClaim(ctx context.Context, tenantRoleID, claimID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roleClaims[tenantRoleID] = append(m.roleClaims[tenantRoleID], claimID)
	return nil
}

func (m *TenantRoleRepositoryMock) RemoveClaim(ctx context.Context, tenantRoleID, claimID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	claims := m.roleClaims[tenantRoleID]
	for i, id := range claims {
		if id == claimID {
			m.roleClaims[tenantRoleID] = append(claims[:i], claims[i+1:]...)
			break
		}
	}
	return nil
}

func (m *TenantRoleRepositoryMock) GetRoleClaims(ctx context.Context, tenantRoleID uuid.UUID) ([]*entity.Claim, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entity.Claim
	for _, claimID := range m.roleClaims[tenantRoleID] {
		if claim, ok := m.claims[claimID]; ok {
			result = append(result, claim)
		}
	}
	return result, nil
}

func (m *TenantRoleRepositoryMock) GetTenantAvailableClaims(ctx context.Context, tenantID uuid.UUID) ([]*entity.Claim, error) {
	return []*entity.Claim{}, nil
}

func (m *TenantRoleRepositoryMock) BatchUpdateClaims(ctx context.Context, tenantRoleID uuid.UUID, assignClaimIDs, removeClaimIDs []uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range removeClaimIDs {
		claims := m.roleClaims[tenantRoleID]
		for i, claimID := range claims {
			if claimID == id {
				m.roleClaims[tenantRoleID] = append(claims[:i], claims[i+1:]...)
				break
			}
		}
	}
	m.roleClaims[tenantRoleID] = append(m.roleClaims[tenantRoleID], assignClaimIDs...)
	return nil
}

func (m *TenantRoleRepositoryMock) AssignToUser(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := userID.String() + ":" + tenantID.String()
	m.userTenantRoles[key] = append(m.userTenantRoles[key], tenantRoleID)
	return nil
}

func (m *TenantRoleRepositoryMock) RemoveFromUser(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := userID.String() + ":" + tenantID.String()
	roles := m.userTenantRoles[key]
	for i, id := range roles {
		if id == tenantRoleID {
			m.userTenantRoles[key] = append(roles[:i], roles[i+1:]...)
			break
		}
	}
	return nil
}

func (m *TenantRoleRepositoryMock) RemoveAllFromUser(ctx context.Context, userID, tenantID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := userID.String() + ":" + tenantID.String()
	delete(m.userTenantRoles, key)
	return nil
}

func (m *TenantRoleRepositoryMock) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.TenantRole, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := userID.String() + ":" + tenantID.String()
	var result []*entity.TenantRole
	for _, roleID := range m.userTenantRoles[key] {
		if role, ok := m.roles[roleID]; ok {
			result = append(result, role)
		}
	}
	return result, nil
}

func (m *TenantRoleRepositoryMock) GetUserClaims(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.Claim, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := userID.String() + ":" + tenantID.String()
	claimMap := make(map[uuid.UUID]*entity.Claim)
	for _, roleID := range m.userTenantRoles[key] {
		for _, claimID := range m.roleClaims[roleID] {
			if claim, ok := m.claims[claimID]; ok {
				claimMap[claim.ID] = claim
			}
		}
	}
	var result []*entity.Claim
	for _, claim := range claimMap {
		result = append(result, claim)
	}
	return result, nil
}

func (m *TenantRoleRepositoryMock) CreateDefaultRoles(ctx context.Context, tenantID uuid.UUID) error {
	return nil
}

func (m *TenantRoleRepositoryMock) AddTenantRole(role *entity.TenantRole) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
}

func (m *TenantRoleRepositoryMock) SetClaim(claim *entity.Claim) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.claims[claim.ID] = claim
}
