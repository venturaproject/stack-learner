package persistence

import (
	"context"
	"errors"
	"time"

	appauth "boilerplate-learner/internal/application/auth"
	domainauth "boilerplate-learner/internal/domain/auth"
	"boilerplate-learner/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

var (
	_ domainauth.Repository  = (*AuthRepository)(nil)
	_ appauth.ReadRepository = (*AuthRepository)(nil)
)

// ── User ────────────────────────────────────────────────────────────────────

func (r *AuthRepository) FindUserByID(ctx context.Context, id int32) (*domainauth.User, error) {
	var m models.User
	if err := connFromContext(ctx, r.db).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapUser(&m), nil
}

// LockUser serializes session refresh against account status changes. Call it
// inside a transaction before consuming or creating refresh tokens.
func (r *AuthRepository) LockUser(ctx context.Context, id int32) (*domainauth.User, error) {
	var m models.User
	err := connFromContext(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"}).First(&m, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapUser(&m), nil
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (*domainauth.User, error) {
	var m models.User
	if err := connFromContext(ctx, r.db).Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapUser(&m), nil
}

func (r *AuthRepository) FindUserByEmailOrUsername(ctx context.Context, login string) (*domainauth.User, error) {
	var m models.User
	err := connFromContext(ctx, r.db).
		Where("email = ? OR username = ?", login, login).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapUser(&m), nil
}

func (r *AuthRepository) LoadUserWithDetails(ctx context.Context, userID int32) (*appauth.UserWithDetails, error) {
	var m models.User
	err := connFromContext(ctx, r.db).
		Preload("Roles.Permissions").
		Preload("Permissions").
		Preload("Settings").
		First(&m, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	roleNames := make([]string, 0, len(m.Roles))
	permSet := map[string]struct{}{}
	for _, rol := range m.Roles {
		roleNames = append(roleNames, rol.Name)
		for _, p := range rol.Permissions {
			permSet[p.Name] = struct{}{}
		}
	}
	for _, p := range m.Permissions {
		permSet[p.Name] = struct{}{}
	}
	perms := make([]string, 0, len(permSet))
	for p := range permSet {
		perms = append(perms, p)
	}

	var avatar *string
	if m.Settings != nil {
		avatar = m.Settings.Avatar
	}

	u := mapUser(&m)
	return &appauth.UserWithDetails{
		ID:          u.ID,
		Name:        u.Name,
		Email:       u.Email,
		Username:    u.Username,
		Status:      u.Status,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		Roles:       roleNames,
		Permissions: perms,
		Avatar:      avatar,
	}, nil
}

func (r *AuthRepository) UserAuthz(ctx context.Context, userID int32) (*appauth.Authz, error) {
	var m models.User
	err := connFromContext(ctx, r.db).
		Preload("Roles.Permissions").
		Preload("Permissions").
		First(&m, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &appauth.Authz{}, nil
		}
		return nil, err
	}

	roles := make([]string, 0, len(m.Roles))
	permSet := map[string]struct{}{}
	for _, rol := range m.Roles {
		roles = append(roles, rol.Name)
		for _, p := range rol.Permissions {
			permSet[p.Name] = struct{}{}
		}
	}
	for _, p := range m.Permissions {
		permSet[p.Name] = struct{}{}
	}
	perms := make([]string, 0, len(permSet))
	for p := range permSet {
		perms = append(perms, p)
	}
	return &appauth.Authz{Active: m.Status == "active", Roles: roles, Permissions: perms}, nil
}

func (r *AuthRepository) CreateUser(ctx context.Context, u domainauth.User) (*domainauth.User, error) {
	m := &models.User{
		Name:         u.Name,
		Email:        u.Email,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		Status:       u.Status,
	}
	if err := connFromContext(ctx, r.db).Create(m).Error; err != nil {
		return nil, err
	}
	return mapUser(m), nil
}

func (r *AuthRepository) UpdateUser(ctx context.Context, u domainauth.User) (*domainauth.User, error) {
	updates := map[string]interface{}{
		"name":          u.Name,
		"email":         u.Email,
		"username":      u.Username,
		"password_hash": u.PasswordHash,
		"status":        u.Status,
		"updated_at":    time.Now(),
	}
	if err := connFromContext(ctx, r.db).Model(&models.User{}).Where("id = ?", u.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.FindUserByID(ctx, u.ID)
}

func (r *AuthRepository) DeleteUser(ctx context.Context, id int32) error {
	return connFromContext(ctx, r.db).Delete(&models.User{}, id).Error
}

func (r *AuthRepository) ListUsers(ctx context.Context, f appauth.UserListFilters) (*appauth.UserListResult, error) {
	q := connFromContext(ctx, r.db).Model(&models.User{})
	if f.Search != nil && *f.Search != "" {
		like := "%" + *f.Search + "%"
		q = q.Where("name ILIKE ? OR email ILIKE ? OR username ILIKE ?", like, like, like)
	}
	if f.Status != nil && *f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}

	var total int64
	q.Count(&total)

	var ms []models.User
	if err := q.Preload("Roles").
		Offset(int((f.Page - 1) * f.PerPage)).Limit(int(f.PerPage)).
		Find(&ms).Error; err != nil {
		return nil, err
	}

	items := make([]appauth.UserListItem, 0, len(ms))
	for _, m := range ms {
		roles := make([]string, 0, len(m.Roles))
		for _, rol := range m.Roles {
			roles = append(roles, rol.Name)
		}
		items = append(items, appauth.UserListItem{
			ID:        m.ID,
			Name:      m.Name,
			Email:     m.Email,
			Username:  m.Username,
			Status:    m.Status,
			Roles:     roles,
			CreatedAt: m.CreatedAt,
		})
	}

	return &appauth.UserListResult{
		Results: items,
		Total:   total,
		Page:    f.Page,
		PerPage: f.PerPage,
	}, nil
}

func (r *AuthRepository) UpsertUserSettings(ctx context.Context, s domainauth.UserSettings) error {
	return connFromContext(ctx, r.db).
		Where(models.UserSettings{UserID: s.UserID}).
		Assign(models.UserSettings{Avatar: s.Avatar}).
		FirstOrCreate(&models.UserSettings{}).Error
}

func (r *AuthRepository) AssignRolesToUser(ctx context.Context, userID int32, roleIDs []int32) error {
	var user models.User
	if err := connFromContext(ctx, r.db).First(&user, userID).Error; err != nil {
		return err
	}
	var roles []models.Role
	if len(roleIDs) > 0 {
		if err := connFromContext(ctx, r.db).Find(&roles, roleIDs).Error; err != nil {
			return err
		}
	}
	return connFromContext(ctx, r.db).Model(&user).Association("Roles").Replace(roles)
}

// ── Refresh tokens ───────────────────────────────────────────────────────────

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, rt domainauth.RefreshToken) error {
	m := &models.RefreshToken{
		JTI:       rt.JTI,
		UserID:    rt.UserID,
		Revoked:   false,
		ExpiresAt: rt.ExpiresAt,
	}
	return connFromContext(ctx, r.db).Create(m).Error
}

func (r *AuthRepository) FindRefreshTokenByJTI(ctx context.Context, jti string) (*domainauth.RefreshToken, error) {
	var m models.RefreshToken
	err := connFromContext(ctx, r.db).Where("jti = ?", jti).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domainauth.RefreshToken{
		ID:        m.ID,
		JTI:       m.JTI,
		UserID:    m.UserID,
		Revoked:   m.Revoked,
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
	}, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, jti string) error {
	return connFromContext(ctx, r.db).
		Model(&models.RefreshToken{}).
		Where("jti = ?", jti).
		Update("revoked", true).Error
}

// ConsumeRefreshToken atomically revokes a still-valid token and reports whether
// this caller won the race. It must run in the transaction that creates the
// replacement token so a failed rotation does not strand the session.
func (r *AuthRepository) ConsumeRefreshToken(ctx context.Context, jti string, now time.Time) (bool, error) {
	result := connFromContext(ctx, r.db).
		Model(&models.RefreshToken{}).
		Where("jti = ? AND revoked = false AND expires_at > ?", jti, now).
		Update("revoked", true)
	return result.RowsAffected == 1, result.Error
}

func (r *AuthRepository) RevokeAllUserTokens(ctx context.Context, userID int32) error {
	return connFromContext(ctx, r.db).
		Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = false", userID).
		Update("revoked", true).Error
}

// ── Roles ────────────────────────────────────────────────────────────────────

func (r *AuthRepository) ListRoles(ctx context.Context) ([]domainauth.Role, error) {
	var ms []models.Role
	if err := connFromContext(ctx, r.db).Preload("Permissions").Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]domainauth.Role, 0, len(ms))
	for _, m := range ms {
		out = append(out, mapRole(&m))
	}
	return out, nil
}

func (r *AuthRepository) FindRoleByID(ctx context.Context, id int32) (*domainauth.Role, error) {
	var m models.Role
	if err := connFromContext(ctx, r.db).Preload("Permissions").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	role := mapRole(&m)
	return &role, nil
}

func (r *AuthRepository) CreateRole(ctx context.Context, role domainauth.Role) (*domainauth.Role, error) {
	m := &models.Role{Name: role.Name, GuardName: role.GuardName}
	if err := connFromContext(ctx, r.db).Create(m).Error; err != nil {
		return nil, err
	}
	created := mapRole(m)
	return &created, nil
}

func (r *AuthRepository) UpdateRole(ctx context.Context, role domainauth.Role) (*domainauth.Role, error) {
	if err := connFromContext(ctx, r.db).Model(&models.Role{}).
		Where("id = ?", role.ID).
		Update("name", role.Name).Error; err != nil {
		return nil, err
	}
	return r.FindRoleByID(ctx, role.ID)
}

func (r *AuthRepository) DeleteRole(ctx context.Context, id int32) error {
	return connFromContext(ctx, r.db).Delete(&models.Role{}, id).Error
}

// ── Permissions ──────────────────────────────────────────────────────────────

func (r *AuthRepository) ListPermissions(ctx context.Context) ([]domainauth.Permission, error) {
	var ms []models.Permission
	if err := connFromContext(ctx, r.db).Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]domainauth.Permission, 0, len(ms))
	for _, m := range ms {
		out = append(out, domainauth.Permission{
			ID:        m.ID,
			Name:      m.Name,
			GuardName: m.GuardName,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}
	return out, nil
}

func (r *AuthRepository) CreatePermission(ctx context.Context, perm domainauth.Permission) (*domainauth.Permission, error) {
	m := &models.Permission{Name: perm.Name, GuardName: perm.GuardName}
	if err := connFromContext(ctx, r.db).Create(m).Error; err != nil {
		return nil, err
	}
	p := domainauth.Permission{ID: m.ID, Name: m.Name, GuardName: m.GuardName, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
	return &p, nil
}

func (r *AuthRepository) DeletePermission(ctx context.Context, id int32) error {
	return connFromContext(ctx, r.db).Delete(&models.Permission{}, id).Error
}

// ── Mappers ──────────────────────────────────────────────────────────────────

func mapUser(m *models.User) *domainauth.User {
	return &domainauth.User{
		ID:           m.ID,
		Name:         m.Name,
		Email:        m.Email,
		Username:     m.Username,
		PasswordHash: m.PasswordHash,
		Status:       m.Status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func mapRole(m *models.Role) domainauth.Role {
	perms := make([]string, 0, len(m.Permissions))
	for _, p := range m.Permissions {
		perms = append(perms, p.Name)
	}
	return domainauth.Role{
		ID:          m.ID,
		Name:        m.Name,
		GuardName:   m.GuardName,
		Permissions: perms,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
