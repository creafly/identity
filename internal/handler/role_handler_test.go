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
	"github.com/hexaend/identity/internal/domain/entity"
	"github.com/hexaend/identity/internal/domain/service"
	"github.com/hexaend/identity/internal/testutil"
)

type RoleServiceMock struct {
	CreateFunc                func(ctx context.Context, input service.CreateRoleInput) (*entity.Role, error)
	GetByIDFunc               func(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	GetByNameFunc             func(ctx context.Context, name string) (*entity.Role, error)
	UpdateFunc                func(ctx context.Context, id uuid.UUID, input service.UpdateRoleInput) (*entity.Role, error)
	DeleteFunc                func(ctx context.Context, id uuid.UUID) error
	ListFunc                  func(ctx context.Context, offset, limit int) ([]*entity.Role, error)
	AssignToUserFunc          func(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveFromUserFunc        func(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRolesFunc          func(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error)
	AssignDefaultUserRoleFunc func(ctx context.Context, userID uuid.UUID) error
}

func (m *RoleServiceMock) Create(ctx context.Context, input service.CreateRoleInput) (*entity.Role, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}
	return nil, nil
}

func (m *RoleServiceMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *RoleServiceMock) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, name)
	}
	return nil, nil
}

func (m *RoleServiceMock) Update(ctx context.Context, id uuid.UUID, input service.UpdateRoleInput) (*entity.Role, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, input)
	}
	return nil, nil
}

func (m *RoleServiceMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *RoleServiceMock) List(ctx context.Context, offset, limit int) ([]*entity.Role, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, offset, limit)
	}
	return []*entity.Role{}, nil
}

func (m *RoleServiceMock) AssignToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if m.AssignToUserFunc != nil {
		return m.AssignToUserFunc(ctx, userID, roleID)
	}
	return nil
}

func (m *RoleServiceMock) RemoveFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if m.RemoveFromUserFunc != nil {
		return m.RemoveFromUserFunc(ctx, userID, roleID)
	}
	return nil
}

func (m *RoleServiceMock) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error) {
	if m.GetUserRolesFunc != nil {
		return m.GetUserRolesFunc(ctx, userID)
	}
	return []*entity.Role{}, nil
}

func (m *RoleServiceMock) AssignDefaultUserRole(ctx context.Context, userID uuid.UUID) error {
	if m.AssignDefaultUserRoleFunc != nil {
		return m.AssignDefaultUserRoleFunc(ctx, userID)
	}
	return nil
}

func setupRoleRouter(roleSvc service.RoleService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewRoleHandler(roleSvc)

	router.POST("/roles", handler.Create)
	router.GET("/roles/:id", handler.GetByID)
	router.GET("/roles", handler.List)
	router.PUT("/roles/:id", handler.Update)
	router.DELETE("/roles/:id", handler.Delete)
	router.POST("/users/:userId/roles", handler.AssignToUser)
	router.DELETE("/users/:userId/roles/:roleId", handler.RemoveFromUser)
	router.GET("/users/:userId/roles", handler.GetUserRoles)

	return router
}

func TestRoleHandler_Create(t *testing.T) {
	roleSvc := &RoleServiceMock{}
	router := setupRoleRouter(roleSvc)

	t.Run("valid request", func(t *testing.T) {
		role := testutil.NewTestRole()
		roleSvc.CreateFunc = func(ctx context.Context, input service.CreateRoleInput) (*entity.Role, error) {
			return role, nil
		}

		body := `{"name": "admin", "description": "Admin role"}`
		req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Create() status = %d, want %d", w.Code, http.StatusCreated)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		body := `{"description": "Some description"}`
		req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Create() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		roleSvc.CreateFunc = func(ctx context.Context, input service.CreateRoleInput) (*entity.Role, error) {
			return nil, service.ErrRoleAlreadyExists
		}

		body := `{"name": "admin"}`
		req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Create() status = %d, want %d", w.Code, http.StatusConflict)
		}
	})
}

func TestRoleHandler_Update(t *testing.T) {
	roleSvc := &RoleServiceMock{}
	router := setupRoleRouter(roleSvc)

	t.Run("valid update", func(t *testing.T) {
		role := testutil.NewTestRole()
		roleSvc.UpdateFunc = func(ctx context.Context, id uuid.UUID, input service.UpdateRoleInput) (*entity.Role, error) {
			return role, nil
		}

		body := `{"name": "updated-name"}`
		req := httptest.NewRequest(http.MethodPut, "/roles/"+role.ID.String(), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Update() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		roleSvc.UpdateFunc = func(ctx context.Context, id uuid.UUID, input service.UpdateRoleInput) (*entity.Role, error) {
			return nil, service.ErrRoleNotFound
		}

		body := `{"name": "updated-name"}`
		req := httptest.NewRequest(http.MethodPut, "/roles/"+uuid.New().String(), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Update() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		roleSvc.UpdateFunc = func(ctx context.Context, id uuid.UUID, input service.UpdateRoleInput) (*entity.Role, error) {
			return nil, service.ErrRoleAlreadyExists
		}

		body := `{"name": "existing-name"}`
		req := httptest.NewRequest(http.MethodPut, "/roles/"+uuid.New().String(), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Update() status = %d, want %d", w.Code, http.StatusConflict)
		}
	})
}

func TestRoleHandler_GetByID(t *testing.T) {
	roleSvc := &RoleServiceMock{}
	router := setupRoleRouter(roleSvc)

	t.Run("existing role", func(t *testing.T) {
		role := testutil.NewTestRole()
		roleSvc.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
			return role, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/roles/"+role.ID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/roles/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		roleSvc.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
			return nil, service.ErrRoleNotFound
		}

		req := httptest.NewRequest(http.MethodGet, "/roles/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestRoleHandler_List(t *testing.T) {
	roleSvc := &RoleServiceMock{}
	router := setupRoleRouter(roleSvc)

	t.Run("list roles", func(t *testing.T) {
		roles := []*entity.Role{testutil.NewTestRole(), testutil.NewTestRole()}
		roleSvc.ListFunc = func(ctx context.Context, offset, limit int) ([]*entity.Role, error) {
			return roles, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/roles?offset=0&limit=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("List() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if response["roles"] == nil {
			t.Error("List() response missing roles")
		}
	})
}

func TestRoleHandler_Delete(t *testing.T) {
	roleSvc := &RoleServiceMock{}
	router := setupRoleRouter(roleSvc)

	t.Run("delete existing role", func(t *testing.T) {
		roleSvc.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		req := httptest.NewRequest(http.MethodDelete, "/roles/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Delete() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		roleSvc.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
			return service.ErrRoleNotFound
		}

		req := httptest.NewRequest(http.MethodDelete, "/roles/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Delete() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestRoleHandler_AssignToUser(t *testing.T) {
	roleSvc := &RoleServiceMock{}
	router := setupRoleRouter(roleSvc)

	t.Run("valid assignment", func(t *testing.T) {
		roleSvc.AssignToUserFunc = func(ctx context.Context, userID, roleID uuid.UUID) error {
			return nil
		}

		userID := uuid.New()
		roleID := uuid.New()
		body := `{"roleId": "` + roleID.String() + `"}`
		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/roles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("AssignToUser() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid user id", func(t *testing.T) {
		body := `{"roleId": "` + uuid.New().String() + `"}`
		req := httptest.NewRequest(http.MethodPost, "/users/invalid-uuid/roles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("AssignToUser() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("role not found", func(t *testing.T) {
		roleSvc.AssignToUserFunc = func(ctx context.Context, userID, roleID uuid.UUID) error {
			return service.ErrRoleNotFound
		}

		userID := uuid.New()
		roleID := uuid.New()
		body := `{"roleId": "` + roleID.String() + `"}`
		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/roles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("AssignToUser() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestRoleHandler_RemoveFromUser(t *testing.T) {
	roleSvc := &RoleServiceMock{}
	router := setupRoleRouter(roleSvc)

	t.Run("valid removal", func(t *testing.T) {
		roleSvc.RemoveFromUserFunc = func(ctx context.Context, userID, roleID uuid.UUID) error {
			return nil
		}

		userID := uuid.New()
		roleID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.String()+"/roles/"+roleID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("RemoveFromUser() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid user id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/invalid-uuid/roles/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("RemoveFromUser() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid role id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/"+uuid.New().String()+"/roles/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("RemoveFromUser() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestRoleHandler_GetUserRoles(t *testing.T) {
	roleSvc := &RoleServiceMock{}
	router := setupRoleRouter(roleSvc)

	t.Run("get user roles", func(t *testing.T) {
		roles := []*entity.Role{testutil.NewTestRole(), testutil.NewTestRole()}
		roleSvc.GetUserRolesFunc = func(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error) {
			return roles, nil
		}

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String()+"/roles", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetUserRoles() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid user id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid/roles", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("GetUserRoles() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}
