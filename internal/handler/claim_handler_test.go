package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/creafly/identity/internal/domain/entity"
	"github.com/creafly/identity/internal/domain/service"
	"github.com/creafly/identity/internal/testutil"
)

type ClaimServiceMock struct {
	CreateFunc           func(ctx context.Context, input service.CreateClaimInput) (*entity.Claim, error)
	GetByIDFunc          func(ctx context.Context, id uuid.UUID) (*entity.Claim, error)
	GetByValueFunc       func(ctx context.Context, value string) (*entity.Claim, error)
	GetOrCreateFunc      func(ctx context.Context, value string) (*entity.Claim, error)
	DeleteFunc           func(ctx context.Context, id uuid.UUID) error
	ListFunc             func(ctx context.Context, offset, limit int) ([]*entity.Claim, error)
	AssignToUserFunc     func(ctx context.Context, userID, claimID uuid.UUID) error
	RemoveFromUserFunc   func(ctx context.Context, userID, claimID uuid.UUID) error
	GetUserClaimsFunc    func(ctx context.Context, userID uuid.UUID) ([]*entity.Claim, error)
	GetUserAllClaimsFunc func(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) ([]*entity.Claim, error)
	AssignToRoleFunc     func(ctx context.Context, roleID, claimID uuid.UUID) error
	RemoveFromRoleFunc   func(ctx context.Context, roleID, claimID uuid.UUID) error
	GetRoleClaimsFunc    func(ctx context.Context, roleID uuid.UUID) ([]*entity.Claim, error)
}

func (m *ClaimServiceMock) Create(ctx context.Context, input service.CreateClaimInput) (*entity.Claim, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}
	return nil, nil
}
func (m *ClaimServiceMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Claim, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *ClaimServiceMock) GetByValue(ctx context.Context, value string) (*entity.Claim, error) {
	if m.GetByValueFunc != nil {
		return m.GetByValueFunc(ctx, value)
	}
	return nil, nil
}
func (m *ClaimServiceMock) GetOrCreate(ctx context.Context, value string) (*entity.Claim, error) {
	if m.GetOrCreateFunc != nil {
		return m.GetOrCreateFunc(ctx, value)
	}
	return nil, nil
}
func (m *ClaimServiceMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}
func (m *ClaimServiceMock) List(ctx context.Context, offset, limit int) ([]*entity.Claim, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, offset, limit)
	}
	return []*entity.Claim{}, nil
}
func (m *ClaimServiceMock) AssignToUser(ctx context.Context, userID, claimID uuid.UUID) error {
	if m.AssignToUserFunc != nil {
		return m.AssignToUserFunc(ctx, userID, claimID)
	}
	return nil
}
func (m *ClaimServiceMock) RemoveFromUser(ctx context.Context, userID, claimID uuid.UUID) error {
	if m.RemoveFromUserFunc != nil {
		return m.RemoveFromUserFunc(ctx, userID, claimID)
	}
	return nil
}
func (m *ClaimServiceMock) GetUserClaims(ctx context.Context, userID uuid.UUID) ([]*entity.Claim, error) {
	if m.GetUserClaimsFunc != nil {
		return m.GetUserClaimsFunc(ctx, userID)
	}
	return []*entity.Claim{}, nil
}
func (m *ClaimServiceMock) GetUserAllClaims(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) ([]*entity.Claim, error) {
	if m.GetUserAllClaimsFunc != nil {
		return m.GetUserAllClaimsFunc(ctx, userID, tenantID)
	}
	return []*entity.Claim{}, nil
}
func (m *ClaimServiceMock) AssignToRole(ctx context.Context, roleID, claimID uuid.UUID) error {
	if m.AssignToRoleFunc != nil {
		return m.AssignToRoleFunc(ctx, roleID, claimID)
	}
	return nil
}
func (m *ClaimServiceMock) RemoveFromRole(ctx context.Context, roleID, claimID uuid.UUID) error {
	if m.RemoveFromRoleFunc != nil {
		return m.RemoveFromRoleFunc(ctx, roleID, claimID)
	}
	return nil
}
func (m *ClaimServiceMock) GetRoleClaims(ctx context.Context, roleID uuid.UUID) ([]*entity.Claim, error) {
	if m.GetRoleClaimsFunc != nil {
		return m.GetRoleClaimsFunc(ctx, roleID)
	}
	return []*entity.Claim{}, nil
}

type TenantServiceMock struct{}

func (m *TenantServiceMock) Create(ctx context.Context, input service.CreateTenantInput) (*entity.Tenant, error) {
	return nil, nil
}
func (m *TenantServiceMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	return nil, nil
}
func (m *TenantServiceMock) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	return nil, nil
}
func (m *TenantServiceMock) Update(ctx context.Context, id uuid.UUID, input service.UpdateTenantInput) (*entity.Tenant, error) {
	return nil, nil
}
func (m *TenantServiceMock) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *TenantServiceMock) List(ctx context.Context, offset, limit int) ([]*entity.Tenant, error) {
	return nil, nil
}
func (m *TenantServiceMock) AddMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	return nil
}
func (m *TenantServiceMock) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	return nil
}
func (m *TenantServiceMock) GetMembers(ctx context.Context, tenantID uuid.UUID) ([]*entity.User, error) {
	return nil, nil
}
func (m *TenantServiceMock) IsMember(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *TenantServiceMock) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*entity.Tenant, error) {
	return nil, nil
}

func setupClaimRouter(claimSvc service.ClaimService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	tenantSvc := &TenantServiceMock{}
	handler := NewClaimHandler(claimSvc, tenantSvc)

	router.POST("/claims", handler.Create)
	router.GET("/claims/:id", handler.GetByID)
	router.GET("/claims", handler.List)
	router.DELETE("/claims/:id", handler.Delete)

	return router
}

func TestClaimHandler_Create(t *testing.T) {
	claimSvc := &ClaimServiceMock{}
	router := setupClaimRouter(claimSvc)

	t.Run("valid request", func(t *testing.T) {
		claim := testutil.NewTestClaim()
		claimSvc.CreateFunc = func(ctx context.Context, input service.CreateClaimInput) (*entity.Claim, error) {
			return claim, nil
		}

		body := `{"value": "test:permission"}`
		req := httptest.NewRequest(http.MethodPost, "/claims", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Create() status = %d, want %d", w.Code, http.StatusCreated)
		}
	})

	t.Run("missing value", func(t *testing.T) {
		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/claims", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Create() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("duplicate value", func(t *testing.T) {
		claimSvc.CreateFunc = func(ctx context.Context, input service.CreateClaimInput) (*entity.Claim, error) {
			return nil, service.ErrClaimAlreadyExists
		}

		body := `{"value": "test:permission"}`
		req := httptest.NewRequest(http.MethodPost, "/claims", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Create() status = %d, want %d", w.Code, http.StatusConflict)
		}
	})
}

func TestClaimHandler_GetByID(t *testing.T) {
	claimSvc := &ClaimServiceMock{}
	router := setupClaimRouter(claimSvc)

	t.Run("existing claim", func(t *testing.T) {
		claim := testutil.NewTestClaim()
		claimSvc.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Claim, error) {
			return claim, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/claims/"+claim.ID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/claims/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		claimSvc.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Claim, error) {
			return nil, service.ErrClaimNotFound
		}

		req := httptest.NewRequest(http.MethodGet, "/claims/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestClaimHandler_List(t *testing.T) {
	claimSvc := &ClaimServiceMock{}
	router := setupClaimRouter(claimSvc)

	t.Run("list claims", func(t *testing.T) {
		claims := []*entity.Claim{testutil.NewTestClaim(), testutil.NewTestClaim()}
		claimSvc.ListFunc = func(ctx context.Context, offset, limit int) ([]*entity.Claim, error) {
			return claims, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/claims?offset=0&limit=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("List() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if response["claims"] == nil {
			t.Error("List() response missing claims")
		}
	})
}

func TestClaimHandler_Delete(t *testing.T) {
	claimSvc := &ClaimServiceMock{}
	router := setupClaimRouter(claimSvc)

	t.Run("delete existing claim", func(t *testing.T) {
		claimSvc.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		req := httptest.NewRequest(http.MethodDelete, "/claims/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Delete() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		claimSvc.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
			return service.ErrClaimNotFound
		}

		req := httptest.NewRequest(http.MethodDelete, "/claims/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Delete() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
