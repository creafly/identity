package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/domain/service"
	"github.com/creafly/identity/internal/testutil"
	"github.com/creafly/identity/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TenantRoleServiceMock struct {
	CreateFunc             func(ctx context.Context, input service.CreateTenantRoleInput) (*entity.TenantRole, error)
	GetByIDFunc            func(ctx context.Context, id uuid.UUID) (*entity.TenantRole, error)
	GetByNameFunc          func(ctx context.Context, tenantID uuid.UUID, name string) (*entity.TenantRole, error)
	UpdateFunc             func(ctx context.Context, id uuid.UUID, input service.UpdateTenantRoleInput) (*entity.TenantRole, error)
	DeleteFunc             func(ctx context.Context, id uuid.UUID) error
	RestoreFunc            func(ctx context.Context, id uuid.UUID) error
	ListByTenantFunc       func(ctx context.Context, tenantID uuid.UUID, includeDeleted bool) ([]*entity.TenantRole, error)
	AddClaimFunc           func(ctx context.Context, tenantRoleID, claimID uuid.UUID) error
	RemoveClaimFunc        func(ctx context.Context, tenantRoleID, claimID uuid.UUID) error
	GetRoleClaimsFunc      func(ctx context.Context, tenantRoleID uuid.UUID) ([]*entity.Claim, error)
	GetAvailableClaimsFunc func(ctx context.Context, tenantID uuid.UUID) ([]*entity.Claim, error)
	BatchUpdateClaimsFunc  func(ctx context.Context, tenantRoleID uuid.UUID, assignClaimIDs, removeClaimIDs []uuid.UUID) error
	AssignToUserFunc       func(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error
	RemoveFromUserFunc     func(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error
	GetUserRolesFunc       func(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.TenantRole, error)
	GetUserClaimsFunc      func(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.Claim, error)
	CreateDefaultRolesFunc func(ctx context.Context, tenantID uuid.UUID) error
	AssignOwnerRoleFunc    func(ctx context.Context, userID, tenantID uuid.UUID) error
}

func (m *TenantRoleServiceMock) Create(ctx context.Context, input service.CreateTenantRoleInput) (*entity.TenantRole, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}
	return nil, nil
}

func (m *TenantRoleServiceMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.TenantRole, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *TenantRoleServiceMock) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*entity.TenantRole, error) {
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, tenantID, name)
	}
	return nil, nil
}

func (m *TenantRoleServiceMock) Update(ctx context.Context, id uuid.UUID, input service.UpdateTenantRoleInput) (*entity.TenantRole, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, input)
	}
	return nil, nil
}

func (m *TenantRoleServiceMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *TenantRoleServiceMock) Restore(ctx context.Context, id uuid.UUID) error {
	if m.RestoreFunc != nil {
		return m.RestoreFunc(ctx, id)
	}
	return nil
}

func (m *TenantRoleServiceMock) ListByTenant(ctx context.Context, tenantID uuid.UUID, includeDeleted bool) ([]*entity.TenantRole, error) {
	if m.ListByTenantFunc != nil {
		return m.ListByTenantFunc(ctx, tenantID, includeDeleted)
	}
	return []*entity.TenantRole{}, nil
}

func (m *TenantRoleServiceMock) AddClaim(ctx context.Context, tenantRoleID, claimID uuid.UUID) error {
	if m.AddClaimFunc != nil {
		return m.AddClaimFunc(ctx, tenantRoleID, claimID)
	}
	return nil
}

func (m *TenantRoleServiceMock) RemoveClaim(ctx context.Context, tenantRoleID, claimID uuid.UUID) error {
	if m.RemoveClaimFunc != nil {
		return m.RemoveClaimFunc(ctx, tenantRoleID, claimID)
	}
	return nil
}

func (m *TenantRoleServiceMock) GetRoleClaims(ctx context.Context, tenantRoleID uuid.UUID) ([]*entity.Claim, error) {
	if m.GetRoleClaimsFunc != nil {
		return m.GetRoleClaimsFunc(ctx, tenantRoleID)
	}
	return []*entity.Claim{}, nil
}

func (m *TenantRoleServiceMock) GetAvailableClaims(ctx context.Context, tenantID uuid.UUID) ([]*entity.Claim, error) {
	if m.GetAvailableClaimsFunc != nil {
		return m.GetAvailableClaimsFunc(ctx, tenantID)
	}
	return []*entity.Claim{}, nil
}

func (m *TenantRoleServiceMock) BatchUpdateClaims(ctx context.Context, tenantRoleID uuid.UUID, assignClaimIDs, removeClaimIDs []uuid.UUID) error {
	if m.BatchUpdateClaimsFunc != nil {
		return m.BatchUpdateClaimsFunc(ctx, tenantRoleID, assignClaimIDs, removeClaimIDs)
	}
	return nil
}

func (m *TenantRoleServiceMock) AssignToUser(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error {
	if m.AssignToUserFunc != nil {
		return m.AssignToUserFunc(ctx, userID, tenantID, tenantRoleID)
	}
	return nil
}

func (m *TenantRoleServiceMock) RemoveFromUser(ctx context.Context, userID, tenantID, tenantRoleID uuid.UUID) error {
	if m.RemoveFromUserFunc != nil {
		return m.RemoveFromUserFunc(ctx, userID, tenantID, tenantRoleID)
	}
	return nil
}

func (m *TenantRoleServiceMock) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.TenantRole, error) {
	if m.GetUserRolesFunc != nil {
		return m.GetUserRolesFunc(ctx, userID, tenantID)
	}
	return []*entity.TenantRole{}, nil
}

func (m *TenantRoleServiceMock) GetUserClaims(ctx context.Context, userID, tenantID uuid.UUID) ([]*entity.Claim, error) {
	if m.GetUserClaimsFunc != nil {
		return m.GetUserClaimsFunc(ctx, userID, tenantID)
	}
	return []*entity.Claim{}, nil
}

func (m *TenantRoleServiceMock) CreateDefaultRoles(ctx context.Context, tenantID uuid.UUID) error {
	if m.CreateDefaultRolesFunc != nil {
		return m.CreateDefaultRolesFunc(ctx, tenantID)
	}
	return nil
}

func (m *TenantRoleServiceMock) AssignOwnerRole(ctx context.Context, userID, tenantID uuid.UUID) error {
	if m.AssignOwnerRoleFunc != nil {
		return m.AssignOwnerRoleFunc(ctx, userID, tenantID)
	}
	return nil
}

func setupTenantRoleRouter(tenantRoleSvc service.TenantRoleService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewTenantRoleHandler(tenantRoleSvc)

	router.POST("/tenants/:id/roles", handler.Create)
	router.GET("/tenants/:id/roles", handler.List)
	router.GET("/tenants/:id/roles/:roleId", handler.GetByID)
	router.PUT("/tenants/:id/roles/:roleId", handler.Update)
	router.DELETE("/tenants/:id/roles/:roleId", handler.Delete)
	router.POST("/tenants/:id/roles/:roleId/claims", handler.AssignClaim)
	router.DELETE("/tenants/:id/roles/:roleId/claims/:claimId", handler.RemoveClaim)
	router.GET("/tenants/:id/roles/:roleId/claims", handler.GetRoleClaims)
	router.GET("/tenants/:id/available-claims", handler.GetAvailableClaims)
	router.PATCH("/tenants/:id/roles/:roleId/claims", handler.BatchUpdateClaims)

	return router
}

func TestTenantRoleHandler_Create(t *testing.T) {
	svc := &TenantRoleServiceMock{}
	router := setupTenantRoleRouter(svc)

	t.Run("valid request", func(t *testing.T) {
		tenantID := utils.GenerateUUID()
		role := testutil.NewTestTenantRole(tenantID)
		svc.CreateFunc = func(ctx context.Context, input service.CreateTenantRoleInput) (*entity.TenantRole, error) {
			return role, nil
		}

		body := `{"name": "manager", "description": "Manager role"}`
		req := httptest.NewRequest(http.MethodPost, "/tenants/"+tenantID.String()+"/roles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Create() status = %d, want %d", w.Code, http.StatusCreated)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		tenantID := utils.GenerateUUID()
		body := `{"description": "Some description"}`
		req := httptest.NewRequest(http.MethodPost, "/tenants/"+tenantID.String()+"/roles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Create() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		svc.CreateFunc = func(ctx context.Context, input service.CreateTenantRoleInput) (*entity.TenantRole, error) {
			return nil, service.ErrTenantRoleAlreadyExists
		}

		tenantID := utils.GenerateUUID()
		body := `{"name": "manager"}`
		req := httptest.NewRequest(http.MethodPost, "/tenants/"+tenantID.String()+"/roles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Create() status = %d, want %d", w.Code, http.StatusConflict)
		}
	})
}

func TestTenantRoleHandler_GetByID(t *testing.T) {
	svc := &TenantRoleServiceMock{}
	router := setupTenantRoleRouter(svc)

	t.Run("existing role", func(t *testing.T) {
		tenantID := utils.GenerateUUID()
		role := testutil.NewTestTenantRole(tenantID)
		svc.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.TenantRole, error) {
			return role, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/tenants/"+tenantID.String()+"/roles/"+role.ID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.TenantRole, error) {
			return nil, service.ErrTenantRoleNotFound
		}

		tenantID := utils.GenerateUUID()
		roleID := utils.GenerateUUID()
		req := httptest.NewRequest(http.MethodGet, "/tenants/"+tenantID.String()+"/roles/"+roleID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestTenantRoleHandler_List(t *testing.T) {
	svc := &TenantRoleServiceMock{}
	router := setupTenantRoleRouter(svc)

	t.Run("list roles", func(t *testing.T) {
		tenantID := utils.GenerateUUID()
		roles := []*entity.TenantRole{testutil.NewTestTenantRole(tenantID), testutil.NewTestTenantRole(tenantID)}
		svc.ListByTenantFunc = func(ctx context.Context, tid uuid.UUID, includeDeleted bool) ([]*entity.TenantRole, error) {
			return roles, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/tenants/"+tenantID.String()+"/roles", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("List() status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestTenantRoleHandler_Update(t *testing.T) {
	svc := &TenantRoleServiceMock{}
	router := setupTenantRoleRouter(svc)

	t.Run("valid update", func(t *testing.T) {
		tenantID := utils.GenerateUUID()
		role := testutil.NewTestTenantRole(tenantID)
		svc.UpdateFunc = func(ctx context.Context, id uuid.UUID, input service.UpdateTenantRoleInput) (*entity.TenantRole, error) {
			return role, nil
		}

		body := `{"name": "updated-name"}`
		req := httptest.NewRequest(http.MethodPut, "/tenants/"+tenantID.String()+"/roles/"+role.ID.String(), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Update() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc.UpdateFunc = func(ctx context.Context, id uuid.UUID, input service.UpdateTenantRoleInput) (*entity.TenantRole, error) {
			return nil, service.ErrTenantRoleNotFound
		}

		tenantID := utils.GenerateUUID()
		roleID := utils.GenerateUUID()
		body := `{"name": "updated-name"}`
		req := httptest.NewRequest(http.MethodPut, "/tenants/"+tenantID.String()+"/roles/"+roleID.String(), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Update() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestTenantRoleHandler_Delete(t *testing.T) {
	svc := &TenantRoleServiceMock{}
	router := setupTenantRoleRouter(svc)

	t.Run("delete existing role", func(t *testing.T) {
		svc.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		tenantID := utils.GenerateUUID()
		roleID := utils.GenerateUUID()
		req := httptest.NewRequest(http.MethodDelete, "/tenants/"+tenantID.String()+"/roles/"+roleID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Delete() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
			return service.ErrTenantRoleNotFound
		}

		tenantID := utils.GenerateUUID()
		roleID := utils.GenerateUUID()
		req := httptest.NewRequest(http.MethodDelete, "/tenants/"+tenantID.String()+"/roles/"+roleID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Delete() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("cannot delete default role", func(t *testing.T) {
		svc.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
			return service.ErrCannotDeleteDefaultRole
		}

		tenantID := utils.GenerateUUID()
		roleID := utils.GenerateUUID()
		req := httptest.NewRequest(http.MethodDelete, "/tenants/"+tenantID.String()+"/roles/"+roleID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Delete() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestTenantRoleHandler_AssignClaim(t *testing.T) {
	svc := &TenantRoleServiceMock{}
	router := setupTenantRoleRouter(svc)

	t.Run("valid assignment", func(t *testing.T) {
		svc.AddClaimFunc = func(ctx context.Context, tenantRoleID, claimID uuid.UUID) error {
			return nil
		}

		tenantID := utils.GenerateUUID()
		roleID := utils.GenerateUUID()
		claimID := utils.GenerateUUID()
		body := `{"claimId": "` + claimID.String() + `"}`
		req := httptest.NewRequest(http.MethodPost, "/tenants/"+tenantID.String()+"/roles/"+roleID.String()+"/claims", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("AssignClaim() status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestTenantRoleHandler_RemoveClaim(t *testing.T) {
	svc := &TenantRoleServiceMock{}
	router := setupTenantRoleRouter(svc)

	t.Run("valid removal", func(t *testing.T) {
		svc.RemoveClaimFunc = func(ctx context.Context, tenantRoleID, claimID uuid.UUID) error {
			return nil
		}

		tenantID := utils.GenerateUUID()
		roleID := utils.GenerateUUID()
		claimID := utils.GenerateUUID()
		req := httptest.NewRequest(http.MethodDelete, "/tenants/"+tenantID.String()+"/roles/"+roleID.String()+"/claims/"+claimID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("RemoveClaim() status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestTenantRoleHandler_GetRoleClaims(t *testing.T) {
	svc := &TenantRoleServiceMock{}
	router := setupTenantRoleRouter(svc)

	t.Run("get claims", func(t *testing.T) {
		claims := []*entity.Claim{testutil.NewTestClaim(), testutil.NewTestClaim()}
		svc.GetRoleClaimsFunc = func(ctx context.Context, tenantRoleID uuid.UUID) ([]*entity.Claim, error) {
			return claims, nil
		}

		tenantID := utils.GenerateUUID()
		roleID := utils.GenerateUUID()
		req := httptest.NewRequest(http.MethodGet, "/tenants/"+tenantID.String()+"/roles/"+roleID.String()+"/claims", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetRoleClaims() status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestTenantRoleHandler_GetAvailableClaims(t *testing.T) {
	svc := &TenantRoleServiceMock{}
	router := setupTenantRoleRouter(svc)

	t.Run("get available claims", func(t *testing.T) {
		claims := []*entity.Claim{testutil.NewTestClaim(), testutil.NewTestClaim()}
		svc.GetAvailableClaimsFunc = func(ctx context.Context, tenantID uuid.UUID) ([]*entity.Claim, error) {
			return claims, nil
		}

		tenantID := utils.GenerateUUID()
		req := httptest.NewRequest(http.MethodGet, "/tenants/"+tenantID.String()+"/available-claims", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetAvailableClaims() status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestTenantRoleHandler_BatchUpdateClaims(t *testing.T) {
	svc := &TenantRoleServiceMock{}
	router := setupTenantRoleRouter(svc)

	t.Run("valid batch update", func(t *testing.T) {
		svc.BatchUpdateClaimsFunc = func(ctx context.Context, tenantRoleID uuid.UUID, assignClaimIDs, removeClaimIDs []uuid.UUID) error {
			return nil
		}

		tenantID := utils.GenerateUUID()
		roleID := utils.GenerateUUID()
		claimID1 := utils.GenerateUUID()
		claimID2 := utils.GenerateUUID()
		body := `{"assignClaimIds": ["` + claimID1.String() + `"], "removeClaimIds": ["` + claimID2.String() + `"]}`
		req := httptest.NewRequest(http.MethodPatch, "/tenants/"+tenantID.String()+"/roles/"+roleID.String()+"/claims", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("BatchUpdateClaims() status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}
