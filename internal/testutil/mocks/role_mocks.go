package mocks

import (
	"context"
	"database/sql"
	"sync"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/google/uuid"
)

type RoleRepositoryMock struct {
	mu        sync.RWMutex
	roles     map[uuid.UUID]*entity.Role
	userRoles map[uuid.UUID][]uuid.UUID
}

func NewRoleRepositoryMock() *RoleRepositoryMock {
	return &RoleRepositoryMock{
		roles:     make(map[uuid.UUID]*entity.Role),
		userRoles: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *RoleRepositoryMock) Create(ctx context.Context, role *entity.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
	return nil
}

func (m *RoleRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	role, ok := m.roles[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return role, nil
}

func (m *RoleRepositoryMock) GetByIDIncludeDeleted(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	role, ok := m.roles[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return role, nil
}

func (m *RoleRepositoryMock) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, role := range m.roles {
		if role.Name == name {
			return role, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *RoleRepositoryMock) Update(ctx context.Context, role *entity.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
	return nil
}

func (m *RoleRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.roles, id)
	return nil
}

func (m *RoleRepositoryMock) Restore(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *RoleRepositoryMock) List(ctx context.Context, offset, limit int, includeDeleted bool) ([]*entity.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var roles []*entity.Role
	for _, role := range m.roles {
		roles = append(roles, role)
	}
	if offset >= len(roles) {
		return []*entity.Role{}, nil
	}
	end := offset + limit
	if end > len(roles) {
		end = len(roles)
	}
	return roles[offset:end], nil
}

func (m *RoleRepositoryMock) AssignToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userRoles[userID] = append(m.userRoles[userID], roleID)
	return nil
}

func (m *RoleRepositoryMock) RemoveFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	roles := m.userRoles[userID]
	for i, id := range roles {
		if id == roleID {
			m.userRoles[userID] = append(roles[:i], roles[i+1:]...)
			break
		}
	}
	return nil
}

func (m *RoleRepositoryMock) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entity.Role
	for _, roleID := range m.userRoles[userID] {
		if role, ok := m.roles[roleID]; ok {
			result = append(result, role)
		}
	}
	return result, nil
}

func (m *RoleRepositoryMock) AddRole(role *entity.Role) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
}
