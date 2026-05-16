package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/username/parking-service/internal/model"
	"github.com/username/parking-service/internal/repository"
	"github.com/username/parking-service/internal/validator"
)

type UserService struct {
	users userStore
}

type UpdateProfileInput struct {
	Name  string  `json:"name"`
	Phone *string `json:"phone"`
}

type UpdateUserInput struct {
	Name  string         `json:"name"`
	Phone *string        `json:"phone"`
	Role  model.UserRole `json:"role"`
}

func NewUserService(users userStore) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Get(ctx context.Context, id uuid.UUID) (model.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return model.User{}, ErrNotFound
	}
	return user, err
}

func (s *UserService) List(ctx context.Context, limit int, offset int) ([]model.User, error) {
	return s.users.List(ctx, limit, offset)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (model.User, error) {
	current, err := s.Get(ctx, userID)
	if err != nil {
		return model.User{}, err
	}
	return s.update(ctx, userID, UpdateUserInput{Name: input.Name, Phone: input.Phone, Role: current.Role})
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (model.User, error) {
	return s.update(ctx, id, input)
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.users.Delete(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *UserService) update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (model.User, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return model.User{}, ValidationError("name", "name is required")
	}
	if input.Phone != nil && !validator.ValidatePhone(*input.Phone) {
		return model.User{}, ValidationError("phone", "invalid phone")
	}
	if input.Role == "" {
		input.Role = model.RoleUser
	}
	if input.Role != model.RoleUser && input.Role != model.RoleOwner && input.Role != model.RoleAdmin {
		return model.User{}, ValidationError("role", "invalid role")
	}

	user, err := s.users.Update(ctx, id, repository.UpdateUserParams{
		Name:  input.Name,
		Phone: input.Phone,
		Role:  input.Role,
	})
	if errors.Is(err, repository.ErrNotFound) {
		return model.User{}, ErrNotFound
	}
	return user, err
}
