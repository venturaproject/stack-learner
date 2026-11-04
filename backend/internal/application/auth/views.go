package auth

import (
	"context"
	"time"

	domainauth "boilerplate-learner/internal/domain/auth"
)

// ── View models (HTTP JSON contract) ─────────────────────────────────────────

type UserWithDetails struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Username    *string   `json:"username"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	Avatar      *string   `json:"avatar"`
}

type UserListItem struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Username  *string   `json:"username"`
	Status    string    `json:"status"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
}

type UserListResult struct {
	Results []UserListItem `json:"results"`
	Total   int64          `json:"count"`
	Page    int64          `json:"page"`
	PerPage int64          `json:"per_page"`
}

// UserListFilters is a query input (not serialized).
type UserListFilters struct {
	Search  *string
	Status  *string
	Page    int64
	PerPage int64
}

type RoleView struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	GuardName   string    `json:"guard_name"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PermissionView struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	GuardName string    `json:"guard_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func roleView(r domainauth.Role) RoleView {
	return RoleView{
		ID: r.ID, Name: r.Name, GuardName: r.GuardName,
		Permissions: r.Permissions, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func roleViews(rs []domainauth.Role) []RoleView {
	out := make([]RoleView, 0, len(rs))
	for _, r := range rs {
		out = append(out, roleView(r))
	}
	return out
}

func permissionView(p domainauth.Permission) PermissionView {
	return PermissionView{
		ID: p.ID, Name: p.Name, GuardName: p.GuardName,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func permissionViews(ps []domainauth.Permission) []PermissionView {
	out := make([]PermissionView, 0, len(ps))
	for _, p := range ps {
		out = append(out, permissionView(p))
	}
	return out
}

// Authz is the effective role/permission set of a user, used by the HTTP
// permission middleware.
type Authz struct {
	Active      bool
	Roles       []string
	Permissions []string
}

// ReadRepository is the query-side port for auth read models. Implemented by
// persistence; not part of the write-side domain Repository.
type ReadRepository interface {
	LoadUserWithDetails(ctx context.Context, userID int32) (*UserWithDetails, error)
	ListUsers(ctx context.Context, f UserListFilters) (*UserListResult, error)
	// UserAuthz returns the user's role names and effective permission names
	// (role permissions ∪ direct permissions).
	UserAuthz(ctx context.Context, userID int32) (*Authz, error)
}
