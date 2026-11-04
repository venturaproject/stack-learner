package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	appauth "boilerplate-learner/internal/application/auth"
	domainauth "boilerplate-learner/internal/domain/auth"
	"boilerplate-learner/internal/infrastructure/http/httpx"
	"boilerplate-learner/internal/shared/apperrors"
)

const (
	UserIDKey      = "user_id"
	rolesKey       = "auth_roles"
	permissionsKey = "auth_permissions"
	fullAccessKey  = "auth_full_access"
)

// Auth validates the access token and loads the caller's roles/permissions into
// the context so RequirePermission can enforce RBAC downstream.
func Auth(
	jwt *appauth.JWTService,
	revoked domainauth.RevokedTokenRepository,
	authz appauth.ReadRepository,
	fullAccessRoles []string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := tokenFromRequest(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
			return
		}

		claims, err := jwt.ParseAccessToken(token)
		if err != nil {
			httpx.Abort(c, err)
			return
		}

		isRevoked, err := revoked.IsRevoked(c.Request.Context(), claims.ID)
		if err != nil || isRevoked {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token revocado"})
			return
		}

		c.Set(UserIDKey, claims.Sub)

		az, err := authz.UserAuthz(c.Request.Context(), claims.Sub)
		if err != nil {
			httpx.Abort(c, apperrors.Internal("no se pudieron cargar los permisos"))
			return
		}
		if !az.Active {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Cuenta desactivada"})
			return
		}
		c.Set(rolesKey, az.Roles)
		c.Set(permissionsKey, az.Permissions)
		c.Set(fullAccessKey, hasFullAccessRole(az.Roles, fullAccessRoles))

		c.Next()
	}
}

// RequirePermission aborts with 403 unless the caller has the given permission
// (or a full-access role). Must run after Auth.
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if HasFullAccess(c) || HasPermission(c, permission) {
			c.Next()
			return
		}
		httpx.Abort(c, apperrors.Forbidden("no tienes permiso para realizar esta acción"))
	}
}

func tokenFromRequest(c *gin.Context) string {
	if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
		return cookie
	}
	return ""
}

func GetUserID(c *gin.Context) (int32, bool) {
	v, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	id, ok := v.(int32)
	return id, ok
}

// HasFullAccess reports whether the caller holds a configured full-access role.
func HasFullAccess(c *gin.Context) bool {
	v, ok := c.Get(fullAccessKey)
	if !ok {
		return false
	}
	full, _ := v.(bool)
	return full
}

// HasPermission reports whether the caller's effective permissions include p.
func HasPermission(c *gin.Context, p string) bool {
	v, ok := c.Get(permissionsKey)
	if !ok {
		return false
	}
	perms, _ := v.([]string)
	return slices.Contains(perms, p)
}

func hasFullAccessRole(roles, fullAccess []string) bool {
	for _, r := range roles {
		if slices.Contains(fullAccess, r) {
			return true
		}
	}
	return false
}
