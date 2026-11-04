package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"boilerplate-learner/internal/infrastructure/http/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func csrfApp() *gin.Engine {
	r := gin.New()
	r.Use(middleware.CSRFGuard())
	r.POST("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func doPost(r *gin.Engine, cookie, csrfHeader string) int {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if csrfHeader != "" {
		req.Header.Set("X-CSRFToken", csrfHeader)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestCSRF_UnauthenticatedPostPasses(t *testing.T) {
	// No access_token → login-style request → CSRF not enforced
	if code := doPost(csrfApp(), "", ""); code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestCSRF_AuthenticatedWithoutCsrfCookieIsRejected(t *testing.T) {
	// access_token present but no csrftoken cookie
	if code := doPost(csrfApp(), "access_token=sometoken", "any"); code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestCSRF_AuthenticatedWithMismatchedTokenIsRejected(t *testing.T) {
	cookie := "access_token=sometoken; csrftoken=correct"
	if code := doPost(csrfApp(), cookie, "wrong"); code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestCSRF_AuthenticatedWithMatchingTokenPasses(t *testing.T) {
	cookie := "access_token=sometoken; csrftoken=abc123"
	if code := doPost(csrfApp(), cookie, "abc123"); code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestCSRF_GetRequestAlwaysPasses(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CSRFGuard())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Cookie", "access_token=sometoken")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
