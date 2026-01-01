package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hexaend/identity/internal/domain/entity"
	"github.com/hexaend/identity/internal/domain/service"
	"github.com/hexaend/identity/internal/testutil"
)

type UserServiceMock struct {
	RegisterFunc            func(ctx context.Context, input service.RegisterInput) (*entity.User, error)
	GetByIDFunc             func(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmailFunc          func(ctx context.Context, email string) (*entity.User, error)
	GetByUsernameFunc       func(ctx context.Context, username string) (*entity.User, error)
	UpdateFunc              func(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*entity.User, error)
	BlockUserFunc           func(ctx context.Context, id uuid.UUID, reason string, blockedBy uuid.UUID) error
	UnblockUserFunc         func(ctx context.Context, id uuid.UUID) error
	DeleteFunc              func(ctx context.Context, id uuid.UUID) error
	ListFunc                func(ctx context.Context, offset, limit int) ([]*entity.User, error)
	ValidateCredentialsFunc func(ctx context.Context, email, password string) (*entity.User, error)
	ChangePasswordFunc      func(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error
}

func (m *UserServiceMock) Register(ctx context.Context, input service.RegisterInput) (*entity.User, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, input)
	}
	return nil, nil
}

func (m *UserServiceMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *UserServiceMock) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *UserServiceMock) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	if m.GetByUsernameFunc != nil {
		return m.GetByUsernameFunc(ctx, username)
	}
	return nil, nil
}

func (m *UserServiceMock) Update(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*entity.User, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, input)
	}
	return nil, nil
}

func (m *UserServiceMock) BlockUser(ctx context.Context, id uuid.UUID, reason string, blockedBy uuid.UUID) error {
	if m.BlockUserFunc != nil {
		return m.BlockUserFunc(ctx, id, reason, blockedBy)
	}
	return nil
}

func (m *UserServiceMock) UnblockUser(ctx context.Context, id uuid.UUID) error {
	if m.UnblockUserFunc != nil {
		return m.UnblockUserFunc(ctx, id)
	}
	return nil
}

func (m *UserServiceMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *UserServiceMock) List(ctx context.Context, offset, limit int) ([]*entity.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, offset, limit)
	}
	return []*entity.User{}, nil
}

func (m *UserServiceMock) ValidateCredentials(ctx context.Context, email, password string) (*entity.User, error) {
	if m.ValidateCredentialsFunc != nil {
		return m.ValidateCredentialsFunc(ctx, email, password)
	}
	return nil, nil
}

func (m *UserServiceMock) ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	if m.ChangePasswordFunc != nil {
		return m.ChangePasswordFunc(ctx, id, oldPassword, newPassword)
	}
	return nil
}

func setupUserRouter(userSvc service.UserService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewUserHandler(userSvc)

	router.GET("/users", handler.List)
	router.GET("/users/:userId", handler.GetByID)
	router.POST("/users/:userId/block", handler.Block)
	router.POST("/users/:userId/unblock", handler.Unblock)

	return router
}

func TestUserHandler_List(t *testing.T) {
	svc := &UserServiceMock{}
	router := setupUserRouter(svc)

	t.Run("list users", func(t *testing.T) {
		users := []*entity.User{testutil.NewTestUser(), testutil.NewTestUser()}
		svc.ListFunc = func(ctx context.Context, offset, limit int) ([]*entity.User, error) {
			return users, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/users?offset=0&limit=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("List() status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestUserHandler_GetByID(t *testing.T) {
	svc := &UserServiceMock{}
	router := setupUserRouter(svc)

	t.Run("existing user", func(t *testing.T) {
		user := testutil.NewTestUser()
		svc.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return user, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return nil, service.ErrUserNotFound
		}

		req := httptest.NewRequest(http.MethodGet, "/users/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestUserHandler_Block(t *testing.T) {
	svc := &UserServiceMock{}
	router := setupUserRouter(svc)

	t.Run("missing userID in context", func(t *testing.T) {
		userID := uuid.New()
		body := `{"reason": "Spam"}`
		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/block", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Block() status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("missing reason", func(t *testing.T) {
		userID := uuid.New()
		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/block", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Block() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestUserHandler_Unblock(t *testing.T) {
	svc := &UserServiceMock{}
	router := setupUserRouter(svc)

	t.Run("valid unblock", func(t *testing.T) {
		svc.UnblockUserFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/unblock", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Unblock() status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc.UnblockUserFunc = func(ctx context.Context, id uuid.UUID) error {
			return service.ErrUserNotFound
		}

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/unblock", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Unblock() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
