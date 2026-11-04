package auth

import (
	"context"
	"time"
)

type Repository interface {
	FindUserByID(ctx context.Context, id int32) (*User, error)
	LockUser(ctx context.Context, id int32) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByEmailOrUsername(ctx context.Context, login string) (*User, error)
	CreateUser(ctx context.Context, user User) (*User, error)
	UpdateUser(ctx context.Context, user User) (*User, error)
	DeleteUser(ctx context.Context, id int32) error

	UpsertUserSettings(ctx context.Context, s UserSettings) error

	CreateRefreshToken(ctx context.Context, rt RefreshToken) error
	FindRefreshTokenByJTI(ctx context.Context, jti string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, jti string) error
	ConsumeRefreshToken(ctx context.Context, jti string, now time.Time) (bool, error)
	RevokeAllUserTokens(ctx context.Context, userID int32) error

	ListRoles(ctx context.Context) ([]Role, error)
	FindRoleByID(ctx context.Context, id int32) (*Role, error)
	CreateRole(ctx context.Context, role Role) (*Role, error)
	UpdateRole(ctx context.Context, role Role) (*Role, error)
	DeleteRole(ctx context.Context, id int32) error

	ListPermissions(ctx context.Context) ([]Permission, error)
	CreatePermission(ctx context.Context, perm Permission) (*Permission, error)
	DeletePermission(ctx context.Context, id int32) error

	AssignRolesToUser(ctx context.Context, userID int32, roleIDs []int32) error
}
