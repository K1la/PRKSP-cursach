package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/username/parking-service/internal/config"
	"github.com/username/parking-service/internal/model"
	"github.com/username/parking-service/internal/repository"
	"github.com/username/parking-service/internal/service"
)

type authOnlyUsers struct{}

func (authOnlyUsers) Create(context.Context, repository.CreateUserParams) (model.User, error) {
	return model.User{}, repository.ErrNotFound
}
func (authOnlyUsers) GetByEmail(context.Context, string) (model.User, error) {
	return model.User{}, repository.ErrNotFound
}
func (authOnlyUsers) GetByID(context.Context, uuid.UUID) (model.User, error) {
	return model.User{}, repository.ErrNotFound
}
func (authOnlyUsers) List(context.Context, int, int) ([]model.User, error) { return nil, nil }
func (authOnlyUsers) Update(context.Context, uuid.UUID, repository.UpdateUserParams) (model.User, error) {
	return model.User{}, repository.ErrNotFound
}
func (authOnlyUsers) Delete(context.Context, uuid.UUID) error { return repository.ErrNotFound }

func TestRequireAuthRejectsInvalidJWT(t *testing.T) {
	api := NewAPI(service.NewAuthService(authOnlyUsers{}, config.Config{JWTSecret: "secret"}), nil, nil, nil, nil, nil)
	handler := api.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRequireRoleRejectsWrongRole(t *testing.T) {
	api := &API{}
	handler := api.RequireRole(model.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), actorContextKey, service.Claims{UserID: uuid.New(), Role: model.RoleOwner})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req.WithContext(ctx))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestRequireAuthAcceptsValidJWT(t *testing.T) {
	secret := "secret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, service.Claims{
		UserID: uuid.New(),
		Role:   model.RoleUser,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	api := NewAPI(service.NewAuthService(authOnlyUsers{}, config.Config{JWTSecret: secret}), nil, nil, nil, nil, nil)
	handler := api.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := ActorFromContext(r.Context()); !ok {
			t.Fatal("actor missing from context")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

func FuzzDecodeJSON(f *testing.F) {
	f.Add([]byte(`{"name":"test"}`))
	f.Add([]byte(``))
	f.Add([]byte(`not-json`))
	f.Fuzz(func(t *testing.T, body []byte) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		var dst map[string]any
		_ = decodeJSON(req, &dst)
	})
}
