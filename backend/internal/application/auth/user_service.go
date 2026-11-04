package auth

import (
	"context"
	"time"

	"github.com/alexedwards/argon2id"

	appshared "boilerplate-learner/internal/application/shared"
	domainauth "boilerplate-learner/internal/domain/auth"
	"boilerplate-learner/internal/shared/apperrors"
)

type CreateUserRequest struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Username *string `json:"username"`
	Password string  `json:"password" binding:"required,min=8"`
	Status   *string `json:"status"`
	RoleIDs  []int32 `json:"role_ids"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Password *string `json:"password,omitempty"`
	Status   *string `json:"status"`
	RoleIDs  []int32 `json:"role_ids"`
}

type UserService struct {
	repo domainauth.Repository
	read ReadRepository
	tx   appshared.TxManager
	pub  appshared.EventPublisher
}

func NewUserService(repo domainauth.Repository, read ReadRepository, tx appshared.TxManager, pub appshared.EventPublisher) *UserService {
	return &UserService{repo: repo, read: read, tx: tx, pub: pub}
}

func (s *UserService) ListUsers(ctx context.Context, f UserListFilters) (*UserListResult, error) {
	return s.read.ListUsers(ctx, f)
}

func (s *UserService) GetUser(ctx context.Context, id int32) (*UserWithDetails, error) {
	if _, err := s.repo.FindUserByID(ctx, id); err != nil {
		return nil, apperrors.NotFound("Usuario no encontrado")
	}
	detail, err := s.read.LoadUserWithDetails(ctx, id)
	if err != nil || detail == nil {
		return nil, apperrors.NotFound("Usuario no encontrado")
	}
	return detail, nil
}

func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*UserWithDetails, error) {
	hash, err := argon2id.CreateHash(req.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}
	status := "active"
	if req.Status != nil {
		status = *req.Status
	}
	var createdID int32
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		created, err := s.repo.CreateUser(ctx, domainauth.User{
			Name:         req.Name,
			Email:        req.Email,
			Username:     req.Username,
			PasswordHash: hash,
			Status:       status,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		})
		if err != nil {
			return err
		}
		if len(req.RoleIDs) > 0 {
			if err := s.repo.AssignRolesToUser(ctx, created.ID, req.RoleIDs); err != nil {
				return err
			}
		}
		createdID = created.ID
		return s.pub.Publish(ctx, domainauth.NewUserCreated(created.ID, created.Email))
	})
	if err != nil {
		return nil, err
	}
	return s.read.LoadUserWithDetails(ctx, createdID)
}

func (s *UserService) UpdateUser(ctx context.Context, id int32, req UpdateUserRequest) (*UserWithDetails, error) {
	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil || user == nil {
		return nil, apperrors.NotFound("Usuario no encontrado")
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Username != nil {
		user.Username = req.Username
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.Password != nil {
		hash, err := argon2id.CreateHash(*req.Password, argon2id.DefaultParams)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hash
	}
	user.UpdatedAt = time.Now()
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := s.repo.UpdateUser(ctx, *user); err != nil {
			return err
		}
		if req.Status != nil && user.Status != "active" {
			if err := s.repo.RevokeAllUserTokens(ctx, id); err != nil {
				return err
			}
		}
		if req.RoleIDs != nil {
			if err := s.repo.AssignRolesToUser(ctx, id, req.RoleIDs); err != nil {
				return err
			}
		}
		return s.pub.Publish(ctx, domainauth.NewUserUpdated(id))
	})
	if err != nil {
		return nil, err
	}
	return s.read.LoadUserWithDetails(ctx, id)
}

func (s *UserService) DeleteUser(ctx context.Context, id int32) error {
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.repo.DeleteUser(ctx, id); err != nil {
			return err
		}
		return s.pub.Publish(ctx, domainauth.NewUserDeleted(id))
	})
}
