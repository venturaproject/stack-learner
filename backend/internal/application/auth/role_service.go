package auth

import (
	"context"
	"time"

	domainauth "boilerplate-learner/internal/domain/auth"
	"boilerplate-learner/internal/shared/apperrors"
)

type RoleRequest struct {
	Name          string  `json:"name" binding:"required"`
	PermissionIDs []int32 `json:"permission_ids"`
}

type RoleService struct {
	repo domainauth.Repository
}

func NewRoleService(repo domainauth.Repository) *RoleService {
	return &RoleService{repo: repo}
}

func (s *RoleService) ListRoles(ctx context.Context) ([]RoleView, error) {
	roles, err := s.repo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	return roleViews(roles), nil
}

func (s *RoleService) GetRole(ctx context.Context, id int32) (*RoleView, error) {
	role, err := s.repo.FindRoleByID(ctx, id)
	if err != nil || role == nil {
		return nil, apperrors.NotFound("Rol no encontrado")
	}
	v := roleView(*role)
	return &v, nil
}

func (s *RoleService) CreateRole(ctx context.Context, req RoleRequest) (*RoleView, error) {
	role, err := s.repo.CreateRole(ctx, domainauth.Role{
		Name:      req.Name,
		GuardName: "web",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return nil, err
	}
	v := roleView(*role)
	return &v, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, id int32, req RoleRequest) (*RoleView, error) {
	role, err := s.repo.FindRoleByID(ctx, id)
	if err != nil || role == nil {
		return nil, apperrors.NotFound("Rol no encontrado")
	}
	role.Name = req.Name
	role.UpdatedAt = time.Now()
	updated, err := s.repo.UpdateRole(ctx, *role)
	if err != nil {
		return nil, err
	}
	v := roleView(*updated)
	return &v, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, id int32) error {
	return s.repo.DeleteRole(ctx, id)
}
