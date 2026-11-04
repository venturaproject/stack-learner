package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CSRFGuard implements the double-submit cookie pattern.
// It only enforces the check on authenticated requests (access_token cookie present)
// so unauthenticated mutations like login are not affected.
func CSRFGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		mutating := isMutating(c.Request.Method)
		if !mutating {
			c.Next()
			return
		}

		// Session-creation endpoints never need CSRF protection: they create
		// or restore the session, so there is no prior valid session to guard.
		// This also avoids false 403s when an expired access_token cookie is
		// still present in the browser.
		switch c.FullPath() {
		case "/api/auth/login", "/api/auth/refresh":
			c.Next()
			return
		}

		_, authErr := c.Cookie("access_token")
		if authErr != nil {
			// Unauthenticated request — no session to protect.
			c.Next()
			return
		}

		csrfCookie, err := c.Cookie("csrftoken")
		if err != nil || csrfCookie == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token ausente"})
			return
		}

		headerToken := c.GetHeader("X-CSRFToken")
		if csrfCookie != headerToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token inválido"})
			return
		}

		c.Next()
	}
}

func isMutating(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}
