package auth

import (
	"context"

	domainauth "boilerplate-learner/internal/domain/auth"
)

type CreatePermissionRequest struct {
	Name string `json:"name" binding:"required"`
}

type PermissionService struct {
	repo domainauth.Repository
}

func NewPermissionService(repo domainauth.Repository) *PermissionService {
	return &PermissionService{repo: repo}
}

func (s *PermissionService) ListPermissions(ctx context.Context) ([]PermissionView, error) {
	ps, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	return permissionViews(ps), nil
}

func (s *PermissionService) CreatePermission(ctx context.Context, req CreatePermissionRequest) (*PermissionView, error) {
	p, err := s.repo.CreatePermission(ctx, domainauth.Permission{Name: req.Name, GuardName: "web"})
	if err != nil {
		return nil, err
	}
	v := permissionView(*p)
	return &v, nil
}

func (s *PermissionService) DeletePermission(ctx context.Context, id int32) error {
	return s.repo.DeletePermission(ctx, id)
}
