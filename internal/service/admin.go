package service

import (
	"context"

	"github.com/username/parking-service/internal/model"
	"github.com/username/parking-service/internal/repository"
)

type AdminService struct {
	admin *repository.AdminRepository
}

func NewAdminService(admin *repository.AdminRepository) *AdminService {
	return &AdminService{admin: admin}
}

func (s *AdminService) Stats(ctx context.Context, actor Claims) (repository.Stats, error) {
	if actor.Role != model.RoleAdmin {
		return repository.Stats{}, ErrForbidden
	}
	return s.admin.Stats(ctx)
}
