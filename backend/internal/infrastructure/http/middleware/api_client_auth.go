package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	appclients "boilerplate-learner/internal/application/api_clients"
)

const (
	ApiClientIDKey     = "api_client_id"
	ApiClientScopesKey = "api_client_scopes"
)

// BearerAuth validates the Authorization: Bearer <token> header against the DB.
func BearerAuth(svc *appclients.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		rawToken := strings.TrimPrefix(authHeader, "Bearer ")
		if rawToken == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization format — use Bearer <token>"})
			return
		}

		client, token, err := svc.ValidateAccessToken(c.Request.Context(), rawToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			return
		}
		if client == nil || token == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token desconocido"})
			return
		}
		if !client.Active {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API client revocado"})
			return
		}
		if time.Now().After(token.AccessExpiresAt) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expirado"})
			return
		}

		// Fire-and-forget: update last_used_at without blocking the request.
		go svc.TouchLastUsed(context.Background(), client.ClientID)

		c.Set(ApiClientIDKey, client.ClientID)
		c.Set(ApiClientScopesKey, token.Scopes)
		c.Next()
	}
}

// RequireScope aborts with 403 if the authenticated API client lacks the required scope.
func RequireScope(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, _ := c.Get(ApiClientScopesKey)
		scopes, _ := v.([]string)
		for _, s := range scopes {
			if s == "*" || s == scope {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Scope insuficiente"})
	}
}
