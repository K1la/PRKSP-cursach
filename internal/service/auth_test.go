package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/username/parking-service/internal/config"
	"github.com/username/parking-service/internal/model"
	"github.com/username/parking-service/internal/repository"
)

type fakeUserStore struct {
	byID    map[uuid.UUID]model.User
	byEmail map[string]model.User
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{byID: map[uuid.UUID]model.User{}, byEmail: map[string]model.User{}}
}

func (s *fakeUserStore) Create(_ context.Context, params repository.CreateUserParams) (model.User, error) {
	if _, ok := s.byEmail[params.Email]; ok {
		return model.User{}, repository.ErrConflict
	}
	user := model.User{
		ID:           uuid.New(),
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		Name:         params.Name,
		Phone:        params.Phone,
		Role:         params.Role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	s.byID[user.ID] = user
	s.byEmail[user.Email] = user
	return user, nil
}

func (s *fakeUserStore) GetByEmail(_ context.Context, email string) (model.User, error) {
	user, ok := s.byEmail[email]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	return user, nil
}

func (s *fakeUserStore) GetByID(_ context.Context, id uuid.UUID) (model.User, error) {
	user, ok := s.byID[id]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	return user, nil
}

func (s *fakeUserStore) List(context.Context, int, int) ([]model.User, error) {
	users := make([]model.User, 0, len(s.byID))
	for _, user := range s.byID {
		users = append(users, user)
	}
	return users, nil
}

func (s *fakeUserStore) Update(_ context.Context, id uuid.UUID, params repository.UpdateUserParams) (model.User, error) {
	user, ok := s.byID[id]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	user.Name = params.Name
	user.Phone = params.Phone
	user.Role = params.Role
	s.byID[id] = user
	s.byEmail[user.Email] = user
	return user, nil
}

func (s *fakeUserStore) Delete(_ context.Context, id uuid.UUID) error {
	user, ok := s.byID[id]
	if !ok {
		return repository.ErrNotFound
	}
	delete(s.byID, id)
	delete(s.byEmail, user.Email)
	return nil
}

func testConfig() config.Config {
	return config.Config{
		JWTSecret:     "test-secret",
		JWTAccessTTL:  time.Minute,
		JWTRefreshTTL: time.Hour,
	}
}

func TestAuthRegisterLoginAndRefresh(t *testing.T) {
	users := newFakeUserStore()
	svc := NewAuthService(users, testConfig())

	registered, err := svc.Register(context.Background(), RegisterInput{
		Email:    "USER@example.com",
		Password: "password123",
		Name:     "Test User",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if registered.User.Email != "user@example.com" {
		t.Fatalf("Register() normalized email = %q", registered.User.Email)
	}
	if _, err := svc.ParseToken(registered.AccessToken, accessTokenType); err != nil {
		t.Fatalf("ParseToken(access) error = %v", err)
	}

	loggedIn, err := svc.Login(context.Background(), LoginInput{Email: "user@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	refreshed, err := svc.Refresh(context.Background(), RefreshInput{RefreshToken: loggedIn.RefreshToken})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatal("Refresh() returned empty tokens")
	}
}

func TestAuthRejectsInvalidCredentials(t *testing.T) {
	users := newFakeUserStore()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), 10)
	if err != nil {
		t.Fatal(err)
	}
	user := model.User{ID: uuid.New(), Email: "user@example.com", PasswordHash: string(hash), Role: model.RoleUser}
	users.byID[user.ID] = user
	users.byEmail[user.Email] = user

	svc := NewAuthService(users, testConfig())
	_, err = svc.Login(context.Background(), LoginInput{Email: "user@example.com", Password: "wrong"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Login() error = %v, want ErrUnauthorized", err)
	}
}
