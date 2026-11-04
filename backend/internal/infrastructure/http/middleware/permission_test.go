package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func runWith(t *testing.T, setup func(*gin.Context), mw gin.HandlerFunc) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	setup(c)
	mw(c)
	if !c.IsAborted() {
		c.Status(http.StatusOK)
	}
	return w.Code
}

func TestRequirePermission_Allows(t *testing.T) {
	code := runWith(t, func(c *gin.Context) {
		c.Set(permissionsKey, []string{"telefonos.view", "telefonos.edit"})
		c.Set(fullAccessKey, false)
	}, RequirePermission("telefonos.edit"))
	if code != http.StatusOK {
		t.Fatalf("want 200, got %d", code)
	}
}

func TestRequirePermission_DeniesWithout(t *testing.T) {
	code := runWith(t, func(c *gin.Context) {
		c.Set(permissionsKey, []string{"telefonos.view"})
		c.Set(fullAccessKey, false)
	}, RequirePermission("telefonos.edit"))
	if code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", code)
	}
}

func TestRequirePermission_FullAccessBypasses(t *testing.T) {
	code := runWith(t, func(c *gin.Context) {
		c.Set(permissionsKey, []string{})
		c.Set(fullAccessKey, true)
	}, RequirePermission("anything.at.all"))
	if code != http.StatusOK {
		t.Fatalf("want 200, got %d", code)
	}
}

func TestHasFullAccessRole(t *testing.T) {
	if !hasFullAccessRole([]string{"editor", "admin"}, []string{"admin"}) {
		t.Fatal("admin should match")
	}
	if hasFullAccessRole([]string{"editor"}, []string{"admin"}) {
		t.Fatal("editor should not match")
	}
}
